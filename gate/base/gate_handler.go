// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package basegate

import (
	log "github.com/dming/lodos/log"
	"github.com/dming/lodos/utils/safemap"
	"fmt"
	"reflect"
)

type gateHandler struct {
	//AgentLearner
	//GateHandler
	gate     *Gate
	sessions *safemap.BeeMap //连接列表
}

func testAsAgentLearner() AgentLearner {
	handler := &gateHandler{
		sessions: safemap.NewBeeMap(),
	}
	return handler
}

func NewGateHandler(gate *Gate) GateHandler {
	handler := &gateHandler{
		gate:     gate,
		sessions: safemap.NewBeeMap(),
	}
	return handler
}

//当连接建立  并且MQTT协议握手成功
func (h *gateHandler) Connect(a Agent) {
	if a.GetSession() != nil {
		h.sessions.Set(a.GetSession().GetSessionid(), a)
	}
}

//当连接关闭	或者客户端主动发送MQTT DisConnect命令
func (h *gateHandler) DisConnect(a Agent) {
	if a.GetSession() != nil {
		h.sessions.Delete(a.GetSession().GetSessionid())
	}
}

/**
 *更新整个Session 通常是其他模块拉取最新数据
 */
func (h *gateHandler) Update(Sessionid string) (result Session, err error) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(Agent); !ok {
		err = fmt.Errorf(" In Update, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return nil, err
	}
	result = agent.(Agent).GetSession()
	return
}

/**
 *Bind the session with the the Userid.
 */
func (h *gateHandler) Bind(Sessionid string, Userid string) (result Session, err error) {
	defer func() {
		if r := recover(); r != nil{
			log.Error("Error in Bind [%s]", r)
		}
	}()

	//log.Debug("bind call")
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return nil, err
	} else if _, ok := agent.(Agent); !ok {
		err = fmt.Errorf(" In Bind, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return nil, err
	}
	agent.(Agent).GetSession().SetUserid(Userid)

	if h.gate.storage != nil && agent.(Agent).GetSession().GetUserid() != "" {
		//可以持久化

		//这里是进程不安全的，所以不能直接操作已有的Map，只能复制后再操作，操作完毕后再赋值回去。
		// 引用指针也不行，所以不能使用 :=   --dming --已经删除了该段代码

		//数据持久化，直接调用这个函数，不要自己处理了。
		h.gate.storage.Storage(Userid, agent.(Agent).GetSession().GetSettings())
	}

	result = agent.(Agent).GetSession()
	return result, nil
}

/**
 *UnBind the session with the the Userid.
 */
func (h *gateHandler) UnBind(Sessionid string) (result Session, err error) {
	//log.Debug("UnBind call")
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(Agent); !ok {
		err = fmt.Errorf("In UnBind, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return
	}
	agent.(Agent).GetSession().SetUserid("")
	result = agent.(Agent).GetSession()
	return
}

/**
 *Push the session with the the Userid.
 */
func (h *gateHandler) PushSettings(Sessionid string, Settings map[string]string) (result Session, err error) {
	//log.Debug("Push call")
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return nil, err
	} else if _, ok := agent.(Agent); !ok {
		err = fmt.Errorf("In PushSettings, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return nil, err
	}
	agent.(Agent).GetSession().SetSettings(Settings)
	result = agent.(Agent).GetSession()

	if h.gate.storage != nil && agent.(Agent).GetSession().GetUserid() != "" {
		err := h.gate.storage.Storage(agent.(Agent).GetSession().GetUserid(), agent.(Agent).GetSession().GetSettings())
		if err != nil {
			log.Error("gate session storage failure")
		}
	}

	return
}

/**
 *Set values (one or many) for the session.
 */
func (h *gateHandler) Set(Sessionid string, key string, value string) (result Session, err error) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return nil, err
	} else if _, ok := agent.(Agent); !ok {
		err = fmt.Errorf("In Set, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return nil, err
	}
	//agent.(Agent).GetSession().GetSettings()[key] = value
	agent.(Agent).GetSession().Set(key, value)
	result = agent.(Agent).GetSession()

	if h.gate.storage != nil && agent.(Agent).GetSession().GetUserid() != "" {
		err := h.gate.storage.Storage(agent.(Agent).GetSession().GetUserid(), agent.(Agent).GetSession().GetSettings())
		if err != nil {
			log.Error("gate session storage failure")
		}
	}

	return
}

/**
 *Remove value from the session.
 */
func (h *gateHandler) Remove(Sessionid string, key string) (result Session, err error) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(Agent); !ok {
		err = fmt.Errorf("In Remove, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return
	}
	//delete(agent.(Agent).GetSession().GetSettings(), key)
	agent.(Agent).GetSession().Remove(key)
	result = agent.(Agent).GetSession()

	if h.gate.storage != nil && agent.(Agent).GetSession().GetUserid() != "" {
		err := h.gate.storage.Storage(agent.(Agent).GetSession().GetUserid(), agent.(Agent).GetSession().GetSettings())
		if err != nil {
			log.Error("gate session storage failure")
		}
	}
	return
}

/**
 *Send message to the session.
 */
func (h *gateHandler) Send(Sessionid string, topic string, body []byte) (err error) {
	//log.Debug("Send call")
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(Agent); !ok {
		err = fmt.Errorf("In Send, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return
	}
	err = agent.(Agent).WriteMsg(topic, body)
	return err
}

/**
 *主动关闭连接
 */
func (h *gateHandler) Close(Sessionid string) (err error) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(Agent); !ok {
		err = fmt.Errorf("In Close, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return
	}
	agent.(Agent).Close()
	return nil
}
