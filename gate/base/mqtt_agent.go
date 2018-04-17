package basegate

import (
	"github.com/dming/lodos/network"
	"bufio"
	"runtime"
	"github.com/dming/lodos/gate/base/mqtt"
	"github.com/dming/lodos/conf"
	"encoding/json"
	"strings"
	"github.com/dming/lodos/log"
	"fmt"
	"time"
	"github.com/dming/lodos/utils/uuid"
	"github.com/dming/lodos/gate"
	"github.com/dming/lodos/module"
	"container/list"
	"math/rand"
	"github.com/dming/lodos/rpc/argsutil"
)

/**
type resultInfo struct {
	Err  error      //错误结果 如果为nil表示请求正确
	Result interface{} //结果
}
*/

type agent struct {
	//gate.Agent
	module 							 module.FullModule
	session                          gate.Session
	conn                             network.Conn
	r                                *bufio.Reader
	w                                *bufio.Writer
	gate                             gate.GateInterface
	client                           *mqtt.Client
	isclose                          bool
	last_storage_heartbeat_data_time int64 //上一次发送存储心跳时间
	rev_num							 int64
	send_num			 			 int64
}

func NewMqttAgent(module module.FullModule) *agent {
	a := &agent{
		module: module,
	}
	return a
}

func (a *agent) OnInit(gate gate.GateInterface, conn network.Conn) error {
	a.conn = conn
	a.gate = gate
	a.r = bufio.NewReaderSize(conn, 256)
	a.w = bufio.NewWriterSize(conn, 256)
	a.isclose = false
	a.rev_num = 0
	a.send_num = 0
	return nil
}


//as function init， call in tcp_server and ws_server
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
		log.Error("Receive login pack's type error:%v \n", pack.GetType())
		return
	}
	info, ok := (pack.GetVariable()).(*mqtt.Connect)
	if !ok {
		log.Error("It's not a mqtt connection package.")
		return
	}
	carrier := make(map[string]string)
	if id := info.GetUserName(); *id != ""{
		carrier["username"] = *id
	}
	if pwd := info.GetPassword(); *pwd != ""{
		carrier["password"] = *pwd
	}
	//log.Debug("Read login pack %s %s %s %s",*id,*psw,info.GetProtocol(),info.GetVersion())
	c := mqtt.NewClient(conf.Conf.Mqtt, a, a.r, a.w, a.conn, info.GetKeepAlive())
	a.client = c
	a.session, err = NewSessionByMap(a.module.GetApp(), map[string]interface{}{
		"Sessionid": Get_uuid(),
		"Network":   a.conn.RemoteAddr().Network(),
		"IP":        a.conn.RemoteAddr().String(),
		"Serverid":  a.module.GetServerId(),
		"Settings":  make(map[string]string),
		"Carrier":	 carrier,
	})
	if err != nil{
		log.Error("gate create agent fail, %s", err.Error())
		return
	}
	a.gate.GetAgentLearner().Connect(a) //发送连接成功的事件, 添加到连接列表, 即 gateHandler sessions
	log.Info("Gate create mqtt-agent success, remote IP is %s", a.session.GetIP())

	//回复客户端 CONNECT
	err = mqtt.WritePack(mqtt.GetConnAckPack(0), a.w)
	if err != nil {
		return
	}

	c.Listen_loop() //开始监听,直到连接中断
	return nil
}

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) OnClose() error {
	a.isclose = true
	a.gate.GetAgentLearner().DisConnect(a) //发送连接断开的事件
	return nil
}

func (a *agent) Destroy() {
	a.conn.Destroy()
}

func (a *agent) IsClosed() bool {
	return a.isclose
}

func (a *agent) GetSession() gate.Session {
	return a.session
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

	toResult := func(a *agent, Topic string, Results []interface{}, errStr string) (err error) {
		if Results == nil || len(Results) < 1 {
			br, _ := a.module.GetApp().ProtocolMarshal(nil, errStr)
			return a.WriteMsg(Topic, br.GetData())
		}
		switch v2 := Results[0].(type) {
		case module.ProtocolMarshal:
			return a.WriteMsg(Topic, v2.GetData())
		}
		b, err := a.module.GetApp().ProtocolMarshal(Results, errStr)
		if err != nil {
			log.Error(err.Error())
			br, _ := a.module.GetApp().ProtocolMarshal(nil, err.Error())
			return a.WriteMsg(Topic, br.GetData())
		} else {
			return a.WriteMsg(Topic, b.GetData())
		}
		return
		/*
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
		*/
	}

	//路由服务
	switch pack.GetType() {
	case mqtt.PUBLISH:
		a.rev_num++
		pub := pack.GetVariable().(*mqtt.Publish)
		topics := strings.Split(*pub.GetTopic(), "/")
		var msgid string
		if len(topics) < 2 {
			log.Error("Topic must be [moduleType@moduleID]/[gateHandler] | [moduleType@moduleID]/[gateHandler]/[msgid]")
			return
		} else if len(topics) == 3 {
			msgid = topics[2]
		}
		//var args map[string]interface{}
		var argsType []string = make([]string, 2)
		var args [][]byte = make([][]byte, 2)
		if pub.GetMsg()[0] =='{' && pub.GetMsg()[len(pub.GetMsg())-1] == '}'{ //start from { and end with }..{}
			//尝试解析为json为map
			var obj interface{} // var obj map[string]interface{}
			err := json.Unmarshal(pub.GetMsg(), &obj)
			if err != nil {
				log.Warning("try to unmarshal as json and get error : %s, %s", err.Error(), string(pub.GetMsg()))
				if msgid != "" {
					toResult(a, *pub.GetTopic(), nil, "The JSON format is incorrect")
				}
				return
			}
			argsType[1] = argsutil.MAP
			args[1] = pub.GetMsg()
			log.Info("pub.GetMsg() is %s", pub.GetMsg())
		} else {
			argsType[1] = argsutil.BYTES
			args[1] = pub.GetMsg()
		}
		//
		hash := ""
		if a.session.GetUserid() != "" {
			hash = a.session.GetUserid()
		} else {
			hash = a.module.GetServerId()
		}
		if (a.gate.GetTracingHandler() != nil) && a.gate.GetTracingHandler().OnRequestTracing(a.session, *pub.GetTopic(), pub.GetMsg()) {
			a.session.CreateRootSpan("gate")
		}

		moduleSession, err := a.module.GetApp().GetRouteServer(topics[0], hash)
		if err != nil {
			if msgid != "" {
				toResult(a, *pub.GetTopic(), nil, fmt.Sprintf("Service(type:%s) not found", topics[0]))
			}
			log.Error(err.Error())
			return
		}
		startsWith := strings.HasPrefix(topics[1], "HD_")
		if !startsWith {
			if msgid != "" {
				toResult(a, *pub.GetTopic(), nil, fmt.Sprintf("Method(%s) must begin with 'HD_'", topics[1]))
			}
			log.Error("Method(%s) must begin with 'HD_'", topics[1])
			return
		}

		if msgid != "" { //msgid means the flag(or Cid) of the msg(call)..
			argsType[0] = RPC_PARAM_SESSION_TYPE
			b, err := a.GetSession().Serializable()
			if err != nil {
				log.Error(err.Error())
				return
			}
			args[0] = b
			results, err := moduleSession.CallArgs(topics[1], argsType, args) //在此调用RPC。--dming
			log.Info("call %s, result is %v, err is %v", topics[1], results, err)
			var errStr string
			if err != nil {
				errStr = err.Error()
			} else {
				errStr = ""
			}
			toResult(a, *pub.GetTopic(), results, errStr) //返回结果给客户端
		}else{ // if msgid is "", call no return rpc invoke
			argsType[0] = RPC_PARAM_SESSION_TYPE
			b, err := a.GetSession().Serializable()
			if err != nil {
				log.Error(err.Error())
				return
			}
			args[0] = b

			e := moduleSession.CallArgsNR(topics[1], argsType, args)
			if e != nil {
				log.Warning("Gate RPC", e.Error())
			}
		}

		if a.GetSession().GetUserid() != "" {
			//这个链接已经绑定Userid
			interval := time.Now().UnixNano()/1000 /1000 /1000 - a.last_storage_heartbeat_data_time //单位秒
			if interval > a.gate.GetMinStorageHeartbeat() {
				//如果用户信息存储心跳包的时长已经大于一秒
				if a.gate.GetStorageHandler() != nil {
					a.gate.GetStorageHandler().Heartbeat(a.GetSession().GetUserid())
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
			if interval > a.gate.GetMinStorageHeartbeat() {
				//如果用户信息存储心跳包的时长已经大于60秒
				if a.gate.GetStorageHandler() != nil {
					a.gate.GetStorageHandler().Heartbeat(a.GetSession().GetUserid())
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



func Get_uuid() string {
	return uuid.Rand().Hex()
}

func TransNumToString(num int64) (string, error) {
	var base int64
	base = 62
	baseHex := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	output_list := list.New()
	for num/base != 0 {
		output_list.PushFront(num % base)
		num = num / base
	}
	output_list.PushFront(num % base)
	str := ""
	for iter := output_list.Front(); iter != nil; iter = iter.Next() {
		str = str + string(baseHex[int(iter.Value.(int64))])
	}
	return str, nil
}

func TransStringToNum(str string) (int64, error) {

	return 0, nil
}

// 函　数：生成随机数
// 概　要：
// 参　数：
//      min: 最小值
//      max: 最大值
// 返回值：
//      int64: 生成的随机数
func RandInt64(min, max int64) int64 {
	if min >= max {
		return max
	}
	return rand.Int63n(max-min) + min
}

func TimeNow() time.Time {
	return time.Now()
}
