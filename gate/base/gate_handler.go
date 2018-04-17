package basegate

import (
	log "github.com/dming/lodos/log"
	"fmt"
	"reflect"
	"github.com/dming/lodos/gate"
	"strings"
	"github.com/dming/lodos/utils"
)

type gateHandler struct {
	//AgentLearner
	//GateHandler
	gate     gate.GateInterface
	sessions *utils.BeeMap //连接列表
}

func testAsAgentLearner() gate.AgentLearner {
	handler := &gateHandler{
		sessions: utils.NewBeeMap(),
	}
	return handler
}

func NewGateHandler(gate *Gate) gate.GateHandler {
	handler := &gateHandler{
		gate:     gate,
		sessions: utils.NewBeeMap(),
	}
	return handler
}

//当连接建立  并且MQTT协议握手成功
func (h *gateHandler) Connect(a gate.Agent) {
	if a.GetSession() != nil {
		h.sessions.Set(a.GetSession().GetSessionid(), a)
	}
	if h.gate.GetSessionLearner() != nil {
		h.gate.GetSessionLearner().Connect(a.GetSession())
	}
}

//当连接关闭	或者客户端主动发送MQTT DisConnect命令
func (h *gateHandler) DisConnect(a gate.Agent) {
	if a.GetSession() != nil {
		h.sessions.Delete(a.GetSession().GetSessionid())
	}
	if h.gate.GetSessionLearner() != nil {
		h.gate.GetSessionLearner().DisConnect(a.GetSession())
	}
}

/**
	Bind(Sessionid string, Userid string) (result gate.Session, err error)                      //Bind the session with the the Userid.
	UnBind(Sessionid string) (result gate.Session, err error)                                   //UnBind the session with the the Userid.
	Set(Sessionid string, key string, value string) (result gate.Session, err error)            //Set values (one or many) for the session.
	Remove(Sessionid string, key string) (result gate.Session, err error)                       //Remove value from the session.
	Push(Sessionid string, Settings map[string]string) (result gate.Session, err error) //推送信息给Session
	Send(Sessionid string, topic string, body []byte) (err error)                          //Send message to the session.
	SendBatch(Sessionids string, topic string, body []byte) (int64, error)                //批量发送
	BroadCast(topic string, body []byte) (int64, error)
	//查询某一个userId是否连接中，这里只是查询这一个网关里面是否有userId客户端连接，如果有多个网关就需要遍历了
	IsConnect(Sessionid string, Userid string) (result bool, err string)
	Close(Sessionid string) (err error)                  //主动关闭连接
	Update(Sessionid string) (result gate.Session, err error) //更新整个Session 通常是其他模块拉取最新数据
	OnDestroy()
 */

/**
*Bind the session with the the Userid.
*/
func (h *gateHandler) Bind(Sessionid string, Userid string) (result gate.Session, err error) {
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
	} else if _, ok := agent.(gate.Agent); !ok {
		err = fmt.Errorf(" In Bind, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return nil, err
	}
	agent.(gate.Agent).GetSession().SetUserid(Userid)

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		/** --dming
		这里是进程不安全的，所以不能直接操作已有的Map，只能复制后再操作，操作完毕后再赋值回去。
		引用指针也不行，所以不能使用 :=   --dming --已经删除了该段代码
		数据持久化，直接调用这个函数，不要自己处理了。
		h.gate.GetStorageHandler().Storage(Userid, agent.(gate.Agent).GetSession())
		*/

		//可以持久化
		data, err := h.gate.GetStorageHandler().Query(Userid)
		if err == nil && data != nil {
			//有已持久化的数据,可能是上一次连接保存的
			oldSession, err := h.gate.NewSession(data)
			if err == nil {
				if agent.(gate.Agent).GetSession().GetSettings() == nil {
					agent.(gate.Agent).GetSession().SetSettings(oldSession.GetSettings())
				} else {
					//合并两个map 并且以 agent.(Agent).GetSession().Settings 已有的(新的)优先
					settings := oldSession.GetSettings()
					if settings != nil {
						//tempSettings := agent.(gate.Agent).GetSession().GetSettings()
						for k, v := range agent.(gate.Agent).GetSession().GetSettings() {
							settings[k] = v //overwrite settings on old session.settings
						}
						agent.(gate.Agent).GetSession().SetSettings(settings)
					}
					//数据持久化
					h.gate.GetStorageHandler().Storage(Userid, agent.(gate.Agent).GetSession())
				}
			} else {
				//解析持久化数据失败
				log.Warning("Sesssion Resolve fail %s", err.Error())
			}
		}
	}

	result = agent.(gate.Agent).GetSession()
	return result, nil
}

/**
 *UnBind the session with the the Userid.
 */
func (h *gateHandler) UnBind(Sessionid string) (result gate.Session, err error) {
	//log.Debug("UnBind call")
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(gate.Agent); !ok {
		err = fmt.Errorf("In UnBind, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return
	}
	agent.(gate.Agent).GetSession().SetUserid("")
	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *Set values (one or many) for the session.
 */
func (h *gateHandler) Set(Sessionid string, key string, value string) (result gate.Session, err error) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return nil, err
	} else if _, ok := agent.(gate.Agent); !ok {
		err = fmt.Errorf("In Set, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return nil, err
	}
	//agent.(gate.Agent).GetSession().GetSettings()[key] = value
	agent.(gate.Agent).GetSession().Set(key, value)
	result = agent.(gate.Agent).GetSession()

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession())
		if err != nil {
			log.Error("gate session storageHandler failure")
		}
	}

	return
}

/**
 *Remove value from the session.
 */
func (h *gateHandler) Remove(Sessionid string, key string) (result gate.Session, err error) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(gate.Agent); !ok {
		err = fmt.Errorf("In Remove, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return
	}
	//delete(agent.(gate.Agent).GetSession().GetSettings(), key)
	agent.(gate.Agent).GetSession().Remove(key)
	result = agent.(gate.Agent).GetSession()

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession())
		if err != nil {
			log.Error("gate session storageHandler failure")
		}
	}
	return
}

/**
 *Push the session with the the Userid.
 */
func (h *gateHandler) Push(Sessionid string, Settings map[string]string) (result gate.Session, err error) {
	//log.Debug("Push call")
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return nil, err
	} else if _, ok := agent.(gate.Agent); !ok {
		err = fmt.Errorf("In Push, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return nil, err
	}
	agent.(gate.Agent).GetSession().SetSettings(Settings)
	result = agent.(gate.Agent).GetSession()

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession())
		if err != nil {
			log.Error("gate session storageHandler failure")
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
	} else if _, ok := agent.(gate.Agent); !ok {
		err = fmt.Errorf("In Send, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return
	}
	err = agent.(gate.Agent).WriteMsg(topic, body)
	return err
}

/**
 *Send message to the session.
 */
func (h *gateHandler) SendBatch(Sessionids string, topic string, body []byte) (int64, error) {
	sessionids := strings.Split(Sessionids, ",")
	var count int64 = 0
	var err error
	for _, sessionid := range sessionids {
		agent := h.sessions.Get(sessionid)
		if agent == nil {
			//log.Warning("No Sesssion found")
			continue
		}
		err = agent.(gate.Agent).WriteMsg(topic, body)
		if err != nil {
			log.Warning("WriteMsg error:", err.Error())
		} else {
			count++
		}
	}
	return count, err
}

/**
 *Send message to the session.
 */
func (h *gateHandler) BroadCast(topic string, body []byte) (int64, error) {
	var count int64 = 0
	var err error
	for _, agent := range h.sessions.Items() {
		err = agent.(gate.Agent).WriteMsg(topic, body)
		if err != nil {
			log.Warning("WriteMsg error:", err.Error())
		} else {
			count++
		}
	}
	return count, err
}

/**
 *更新整个Session 通常是其他模块拉取最新数据
 */
func (h *gateHandler) Update(Sessionid string) (result gate.Session, err error) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(gate.Agent); !ok {
		err = fmt.Errorf(" In Update, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return nil, err
	}
	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *查询某一个userId是否连接中，这里只是查询这一个网关里面是否有userId客户端连接，如果有多个网关就需要遍历了
 */
func (h *gateHandler) IsConnect(Sessionid string, Userid string) (result bool, err error) {
	for _, agent := range h.sessions.Items() {
		if agent.(gate.Agent).GetSession().GetUserid() == Userid {
			return !agent.(gate.Agent).IsClosed(), nil
		}
	}

	return false, fmt.Errorf("The gateway did not find the corresponding userId 【%s】", Userid)

}

/**
 *主动关闭连接
 */
func (h *gateHandler) Close(Sessionid string) (err error) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = fmt.Errorf("No Sesssion found")
		return
	} else if _, ok := agent.(gate.Agent); !ok {
		err = fmt.Errorf("In Close, %s can not convert to Agent", reflect.TypeOf(agent).String())
		return
	}
	agent.(gate.Agent).Close()
	return nil
}

/**
 *主动关闭连接
 */
func (h *gateHandler) OnDestroy() {
	for _, v := range h.sessions.Items() {
		v.(gate.Agent).Close()
	}
	h.sessions.DeleteAll()
}




