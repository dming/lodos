package basegate

import (
	"bufio"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/network"
	"time"
	"github.com/dming/lodos/module"
	"fmt"
	"reflect"
	log "github.com/dming/lodos/log"
	"github.com/dming/lodos/module/base"
	"github.com/dming/lodos/gate"
)

var RPC_PARAM_SESSION_TYPE="SESSION"
var RPC_PARAM_ProtocolMarshal_TYPE = "ProtocolMarshal"


type Gate struct {
	//module.RPCSerialize
	basemodule.BaseModule
	MaxConnNum          int
	MaxMsgLen           uint32
	MinStorageHeartbeat int64 //Session持久化最短心跳包

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration

	// tcp
	TCPAddr string

	//tls
	Tls      bool
	CertFile string
	KeyFile  string
	//
	gateHandler    gate.GateHandler
	agentLearner   gate.AgentLearner
	sessionLearner gate.SessionLearner
	storageHandler gate.StorageHandler
	tracingHandler gate.TracingHandler
}

func (gate *Gate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Gate"
}
func (gate *Gate) Version() string {
	//可以在监控时了解代码版本
	return "v0.0.0-unset"
}

func (this *Gate) SetGateHandler(gateHandler gate.GateHandler) error {
	this.gateHandler = gateHandler
	return nil
}
func (this *Gate) SetAgentLearner(agentLearner gate.AgentLearner) error {
	this.agentLearner = agentLearner
	return nil
}
/**
设置客户端连接和断开的监听器
*/
func (this *Gate) SetSessionLearner(sessionLearner gate.SessionLearner) error {
	this.sessionLearner = sessionLearner
	return nil
}
/**
设置Session信息持久化接口
*/
func (this *Gate) SetStorageHandler(storageHandler gate.StorageHandler) error {
	this.storageHandler = storageHandler
	return nil
}
func (this *Gate) SetTracingHandler(tracingHandler gate.TracingHandler) error {
	this.tracingHandler = tracingHandler
	return nil
}


func (this *Gate) GetGateHandler() gate.GateHandler {
	return this.gateHandler
}
func (this *Gate) GetAgentLearner() gate.AgentLearner {
	return this.agentLearner
}
func (this *Gate) GetSessionLearner() gate.SessionLearner {
	return this.sessionLearner
}
func (this *Gate) GetStorageHandler() gate.StorageHandler {
	return this.storageHandler
}
func (this *Gate) GetTracingHandler() gate.TracingHandler {
	return this.tracingHandler
}

func (this *Gate) GetMinStorageHeartbeat() int64 {
	return this.MinStorageHeartbeat
}

func (this *Gate) NewSession(data []byte) (gate.Session, error) {
	return NewSession(this.GetApp(), data)
}
func (this *Gate) NewSessionByMap(data map[string]interface{}) (gate.Session, error) {
	return NewSessionByMap(this.GetApp(), data)
}

/**
自定义rpc参数序列化反序列化  Session
 */
func (this *Gate)Serialize(param interface{}) (ptype string,p []byte, err error){
	switch v2 := param.(type) {
	case gate.Session:
		bytes, err := v2.Serializable()
		if err != nil{
			return RPC_PARAM_SESSION_TYPE, nil, err
		}
		return RPC_PARAM_SESSION_TYPE, bytes, nil
	default:
		return "", nil,fmt.Errorf("args [%s] Types not allowed",reflect.TypeOf(param))
	}
}

func (this *Gate)Deserialize(ptype string,b []byte)(param interface{},err error){
	switch ptype {
	case RPC_PARAM_SESSION_TYPE:
		mps, errs:= NewSession(this.GetApp(), b)
		if errs!=nil{
			return	nil,errs
		}
		return mps,nil
	case RPC_PARAM_ProtocolMarshal_TYPE:
		return this.GetApp().NewProtocolMarshal(b), nil
	default:
		return	nil,fmt.Errorf("args [%s] Types not allowed",ptype)
	}
}

func (this *Gate)GetTypes()([]string){
	return []string{RPC_PARAM_SESSION_TYPE}
}

func (this *Gate) OnInit(app module.AppInterface, settings *conf.ModuleSettings) {
	this.BaseModule.OnInit(app, this, settings) //这是必须的
	//添加Session结构体的序列化操作类
	err := app.AddRPCSerialize("this", this)
	if err!=nil{
		log.Warning("Adding session structures failed to serialize interfaces",err.Error())
	}

	this.MaxConnNum = int(settings.Settings["MaxConnNum"].(float64))
	this.MaxMsgLen = uint32(settings.Settings["MaxMsgLen"].(float64))
	if WSAddr, ok := settings.Settings["WSAddr"]; ok {
		this.WSAddr = WSAddr.(string)
	}
	this.HTTPTimeout = time.Second * time.Duration(settings.Settings["HTTPTimeout"].(float64))
	if TCPAddr, ok := settings.Settings["TCPAddr"]; ok {
		this.TCPAddr = TCPAddr.(string)
	}
	if Tls, ok := settings.Settings["Tls"]; ok {
		this.Tls = Tls.(bool)
	} else {
		this.Tls = false
	}
	if CertFile, ok := settings.Settings["CertFile"]; ok {
		this.CertFile = CertFile.(string)
	} else {
		this.CertFile = ""
	}
	if KeyFile, ok := settings.Settings["KeyFile"]; ok {
		this.KeyFile = KeyFile.(string)
	} else {
		this.KeyFile = ""
	}

	if MinHBStorage, ok := settings.Settings["MinHBStorage"]; ok {
		this.MinStorageHeartbeat = int64(MinHBStorage.(float64))
	} else {
		this.MinStorageHeartbeat = 60
	}

	handler := NewGateHandler(this)

	this.agentLearner = handler.(gate.AgentLearner)
	this.gateHandler = handler

	this.GetServer().GetRpcServer().Register("Update", this.gateHandler.Update)
	this.GetServer().GetRpcServer().Register("Bind", this.gateHandler.Bind)
	this.GetServer().GetRpcServer().Register("UnBind", this.gateHandler.UnBind)
	this.GetServer().GetRpcServer().Register("Push", this.gateHandler.Push)
	this.GetServer().GetRpcServer().Register("Set", this.gateHandler.Set)
	this.GetServer().GetRpcServer().Register("Remove", this.gateHandler.Remove)
	this.GetServer().GetRpcServer().Register("Send", this.gateHandler.Send)
	this.GetServer().GetRpcServer().Register("SendBatch", this.gateHandler.SendBatch)
	this.GetServer().GetRpcServer().Register("BroadCast", this.gateHandler.BroadCast)
	this.GetServer().GetRpcServer().Register("IsConnect", this.gateHandler.IsConnect)
	this.GetServer().GetRpcServer().Register("Close", this.gateHandler.Close)
}

func (this *Gate) Run(closeSig chan bool) {
	var wsServer *network.WSServer
	if this.WSAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = this.WSAddr
		wsServer.MaxConnNum = this.MaxConnNum
		wsServer.MaxMsgLen = this.MaxMsgLen
		wsServer.HTTPTimeout = this.HTTPTimeout
		wsServer.Tls = this.Tls
		wsServer.CertFile = this.CertFile
		wsServer.KeyFile = this.KeyFile
		wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
			a := &agent{
				conn:     conn,
				gate:     this,
				r:        bufio.NewReader(conn),
				w:        bufio.NewWriter(conn),
				isclose:  false,
				rev_num:  0,
				send_num: 0,
			}
			return a
		}
	}

	var tcpServer *network.TCPServer
	if this.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = this.TCPAddr
		tcpServer.MaxConnNum = this.MaxConnNum
		tcpServer.Tls = this.Tls
		tcpServer.CertFile = this.CertFile
		tcpServer.KeyFile = this.KeyFile
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &agent{
				conn:     conn,
				gate:     this,
				r:        bufio.NewReader(conn),
				w:        bufio.NewWriter(conn),
				isclose:  false,
				rev_num:  0,
				send_num: 0,
			}
			return a
		}
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	if this.gateHandler != nil {
		this.gateHandler.OnDestroy()
	}
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (this *Gate) OnDestroy() {
	this.BaseModule.OnDestroy() //这是必须的
}
