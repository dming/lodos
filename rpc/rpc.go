package rpc

import (
	"reflect"
	"github.com/dming/lodos/rpc/pb"
	"github.com/streadway/amqp"
)

type CallInfo struct {
	Id      string        // function name
	Args    []interface{} // arguments
	ChanRet chan *RetInfo // channel for return information
	ReplyTo string // mq name for return information
	Flag string // flag the info
}

type RetInfo struct {
	Ret []interface{} // the return result
	Err error
	ReplyTo string
	Flag string // flag the info
}

type FuncInfo struct {
	Fn reflect.Value
	IsGo bool
}

// interfaces

type Client interface {
	AttachChanCall(chanCall chan *CallInfo)
	AttachMqClient(mqClient MqClient)
	Call(id string, timeout int, args ...interface{}) (*RetInfo , error)
	AsynCall(id string, args...interface{}) (chan *RetInfo, error)
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


