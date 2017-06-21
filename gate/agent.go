package gate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/network/mqtt"
	log "github.com/dming/lodos/mlog"
	"github.com/dming/lodos/network"
	"github.com/dming/lodos/utils/uuid"
	"runtime"
	"strings"
	"time"
)

type resultInfo struct {
	Err  error      //错误结果 如果为nil表示请求正确
	Result interface{} //结果
}

type agent struct {
	Agent
	session                          Session
	conn                             network.Conn
	r                                *bufio.Reader
	w                                *bufio.Writer
	gate                             *Gate
	client                           *mqtt.Client
	isclose                          bool
	last_storage_heartbeat_data_time int64 //上一次发送存储心跳时间
	rev_num				int64
	send_num			int64
}

func (a *agent) IsClosed() bool {
	return a.isclose
}

func (a *agent) GetSession() Session {
	return a.session
}

func (a *agent) Run() (err error) {
	defer func() {
		if err := recover(); err != nil {
			buff := make([]byte, 4096)
			runtime.Stack(buff, false)
			log.Error("conn.serve() panic(%v)\n info:%s", err, string(buff))
		}
		a.Close()

	}()
	//握手协议
	var pack *mqtt.Pack
	pack, err = mqtt.ReadPack(a.r)
	if err != nil {
		log.Error("Read login pack error", err)
		return
	}
	if pack.GetType() != mqtt.CONNECT {
		log.Error("Recive login pack's type error:%v \n", pack.GetType())
		return
	}
	info, ok := (pack.GetVariable()).(*mqtt.Connect)
	if !ok {
		log.Error("It's not a mqtt connection package.")
		return
	}
	//id := info.GetUserName()
	//psw := info.GetPassword()
	//log.Debug("Read login pack %s %s %s %s",*id,*psw,info.GetProtocol(),info.GetVersion())
	c := mqtt.NewClient(conf.Conf.Mqtt, a, a.r, a.w, a.conn, info.GetKeepAlive())
	a.client = c
	a.session,err= NewSessionByMap(a.gate.App, map[string]interface{}{
		"Sessionid": Get_uuid(),
		"Network":   a.conn.RemoteAddr().Network(),
		"IP":        a.conn.RemoteAddr().String(),
		"Serverid":  a.gate.GetServerId(),
		"Settings":  make(map[string]string),
	})
	if err!=nil{
		log.Error("gate create agent fail",err.Error())
		return
	}
	a.gate.agentLearner.Connect(a) //发送连接成功的事件

	//回复客户端 CONNECT
	err = mqtt.WritePack(mqtt.GetConnAckPack(0), a.w)
	if err != nil {
		return
	}

	c.Listen_loop() //开始监听,直到连接中断
	return nil
}

func (a *agent) OnClose() error {
	a.isclose = true
	a.gate.agentLearner.DisConnect(a) //发送连接断开的事件
	return nil
}

func (a *agent)RevNum() int64{
	return a.rev_num
}
func (a *agent)SendNum() int64{
	return a.send_num
}

//路由得到的pack to RPC server and return the result to client
func (a *agent) OnRecover(pack *mqtt.Pack) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Gate  OnRecover error [%s]", r)
		}
	}()

	toResult := func(a *agent, Topic string, Result interface{}, Error error) (err error) {
		r := &resultInfo{
			Err:  Error,
			Result: Result,
		}
		b, err := json.Marshal(r)
		if err == nil {
			a.WriteMsg(Topic, b)
		}else{
			r = &resultInfo{
				Err:  err,
				Result: nil,
			}
			log.Error(err.Error())

			br, _ := json.Marshal(r)
			a.WriteMsg(Topic, br)
		}
		return
	}

	//路由服务
	switch pack.GetType() {
	case mqtt.PUBLISH:
		a.rev_num=a.rev_num+1
		pub := pack.GetVariable().(*mqtt.Publish)
		topics := strings.Split(*pub.GetTopic(), "/")
		var msgid string
		if len(topics) < 2 {
			log.Error("Topic must be [moduleType@moduleID]/[handler] | [moduleType@moduleID]/[handler]/[msgid]")
			return
		} else if len(topics) == 3 {
			msgid = topics[2]
		}
		var args map[string]interface{}
		if pub.GetMsg()[0]=='{'&&pub.GetMsg()[len(pub.GetMsg())-1]=='}'{ //start from { and end with }..{}
			//尝试解析为json为map
			var obj map[string]interface{}
			err := json.Unmarshal(pub.GetMsg(), &obj)
			if err != nil {
				log.Debug("try to unmarshal as json and get error : %s, %s", err.Error(), string(pub.GetMsg()))
				if msgid != "" {
					toResult(a, *pub.GetTopic(), nil, fmt.Errorf("The JSON format is incorrect"))
				}
				return
			}
			args = obj
		}else{
			log.Debug("msg should be unmarshal as json : %s", string(pub.GetMsg()))
			if msgid != "" {
				toResult(a, *pub.GetTopic(), nil, fmt.Errorf("The JSON format is incorrect"))
			}
			return
		}
		hash := ""
		if a.session.GetUserid() != "" {
			hash = a.session.GetUserid()
		} else {
			hash = a.gate.GetServerId()
		}

		moduleSession, err := a.gate.GetApp().GetModuleSession(topics[0], hash)
		if err != nil {
			if msgid != "" {
				toResult(a, *pub.GetTopic(), nil, fmt.Errorf("Service(type:%s) not found", topics[0]))
			}
			return
		}
		startsWith := strings.HasPrefix(topics[1], "HD_")
		if !startsWith {
			if msgid != "" {
				toResult(a, *pub.GetTopic(), nil, fmt.Errorf("Method(%s) must begin with 'HD_'", topics[1]))
			}
			return
		}
		if msgid != "" { //msgid means the flag of the msg(call)
			//...something wrong
			result, err := moduleSession.Call(topics[1], 5, a.GetSession(), args) //在此调用RPC。--dming
			log.Info("call %s, result is %v, err is %v", topics[1], result, err)
			toResult(a, *pub.GetTopic(), result, err) //返回结果给客户端
			//...
		}else{
			_, e := moduleSession.Call(topics[1], 5, a.GetSession(), args)
			if e!=nil{
				log.Warning("Gate RPC",e.Error())
			}
		}

		if a.GetSession().GetUserid() != "" {
			//这个链接已经绑定Userid
			interval := time.Now().UnixNano()/1000 /1000 /1000 - a.last_storage_heartbeat_data_time //单位秒
			if interval > a.gate.MinStorageHeartbeat {
				//如果用户信息存储心跳包的时长已经大于一秒
				if a.gate.storage != nil {
					a.gate.storage.Heartbeat(a.GetSession().GetUserid())
					a.last_storage_heartbeat_data_time = time.Now().UnixNano() / 1000 / 1000 / 1000
				}
			}
		}
	case mqtt.PINGREQ:
		log.Debug("Get mqtt.PINGREQ package")
		//客户端发送的心跳包
		if a.GetSession().GetUserid() != "" {
			//这个链接已经绑定Userid
			interval := time.Now().UnixNano()/1000 /1000 /1000 - a.last_storage_heartbeat_data_time //单位秒
			if interval > a.gate.MinStorageHeartbeat {
				//如果用户信息存储心跳包的时长已经大于60秒
				if a.gate.storage != nil {
					a.gate.storage.Heartbeat(a.GetSession().GetUserid())
					a.last_storage_heartbeat_data_time = time.Now().UnixNano() / 1000 /1000 / 1000
				}
			}
		}
	}
}

func (a *agent) WriteMsg(topic string, body []byte) error {
	a.send_num++
	return a.client.WriteMsg(topic, body)
}

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) Destroy() {
	a.conn.Destroy()
}

func Get_uuid() string {
	return uuid.Rand().Hex()
}
