package rpc

import (
	"github.com/dming/lodos/rpc/pb"
	"github.com/dming/lodos/conf"
)

/*
type CallInfo_old struct {
	Id      string        // function name
	Args    []interface{} // arguments
	ChanRet chan *RetInfo // channel for return information
	ReplyTo string // mq name for return information
	Flag string // flag the info
}
type RetInfo_old struct {
	Ret []interface{} // the return result
	Err error
	ReplyTo string
	Flag string // flag the info
}
*/

type CallInfo struct {
	RpcInfo rpcpb.RPCInfo
	Result rpcpb.ResultInfo
	Props map[string]interface{}
	Agent MQServer
}

type FuncInfo struct {
	//Fn reflect.Value
	Function interface{}
	Goroutine bool
}

// interfaces

type MQServer interface {
	Callback(callInfo CallInfo) (err error)
}

type RPCServer interface {
	NewRabbitmqRpcServer(info *conf.Rabbitmq) (err error)
	NewRedisRpcServer(info *conf.Redis) (err error)
	SetListener(listener RPCListener)
	SetGoroutineControl(control GoroutineControl)
	GetExecuting() int64
	GetLocalRpcServer() LocalServer
	Register(id string, fn interface{})
	RegisterGo(id string, fn interface{})
	Done() (err error)
}

type RPCClient interface {
	NewRabbitmqRpcClient(info *conf.Rabbitmq) (err error)
	NewRedisRpcClient(info *conf.Redis) (err error)
	NewLocalRpcClient(server RPCServer) (err error)
	Done () (err error)
	Call(_func string, params ...interface{}) ([]interface{}, error)
	SyncCall(_func string, params ...interface{}) (chan rpcpb.ResultInfo, error)
	CallNR(_func string, params ...interface{}) (err error)
	CallArgs(_func string, ArgsType []string, Args [][]byte) ([]interface{}, error)
	SyncCallArgs(_func string, ArgsType []string, Args [][]byte) (chan rpcpb.ResultInfo, error)
	CallArgsNR(_func string, ArgsType []string, Args [][]byte) (err error)
}

type LocalServer interface {
	IsClose() bool
	WriteToRpcServer(callInfo CallInfo) (success bool)
	StopConsume() (err error)
	Shutdown() (err error)
	Callback(callInfo CallInfo) (err error)
}

type LocalClient interface {
	Done() (err error)
	Call(callInfo CallInfo, callback chan rpcpb.ResultInfo) (err error)
	CallNR(callInfo CallInfo) (err error)
}

type RPCListener interface {
	//need to be complete
}

type GoroutineControl interface {
	Wait() (err error)
	Finish()
}

/*
type Client interface {
	AttachChanCall(chanCall chan *CallInfo)
	AttachMqClient(mqClient MqClient)
	Call(id string, timeout int, args ...interface{}) (*RetInfo , error)
	AsynCall(id string, args ...interface{}) (chan *RetInfo, error)
	Close() error
}

type Server interface {
	GetChanCall() (chan *CallInfo)
	AttachMqServer(mqServer MqServer)
	Register(id string, f interface{})
	RegisterGo(id string, f interface{})
	Run(chanCall chan *CallInfo, chanDone chan error)
	Exec(ci *CallInfo)
	Close() error
}

type MqClient interface {
	Call(ci *CallInfo) error
	Run(deliveries <-chan amqp.Delivery, done chan error)
	Done() (err error)
	Marshal(ci *CallInfo) ([]byte, error)
	UnmarshalMqRetInfo(data []byte) (*rpcpb.MqRetInfo, error)
}

type MqServer interface {
	Run(chanDeli <-chan amqp.Delivery, done chan error)
	StopConsume() error
	Shutdown() error
	MarshalRetInfo(ri *RetInfo) ([]byte, error)
	UnmarshalMqCallInfo(data []byte) (*rpcpb.MqCallInfo, error)
}
*/

