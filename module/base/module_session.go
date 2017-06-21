package basemodule

import (
	"github.com/dming/lodos/rpc"
	"fmt"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/rpc/base"
)

type moduleSession struct {
	Id string
	App module.AppInterface
	Type string
	rpcClient rpc.Client
}

func NewModuleSession(app module.AppInterface, id string, Type string, chanCall chan *rpc.CallInfo, info *conf.Rabbitmq) (module.ModuleSession, error) {
	m := new(moduleSession)
	m.App = app
	m.Id = id
	m.Type = Type
	err := m.CreateClient(chanCall, info)
	return m, err
}


func (m *moduleSession) CreateClient(chanCall chan *rpc.CallInfo, info *conf.Rabbitmq) error {
	var err error
	m.rpcClient = baserpc.NewClient()
	if err != nil {
		fmt.Printf("error in moduleSession.CreateClient, %s", err.Error())
		return err
	}
	if chanCall != nil {
		m.rpcClient.AttachChanCall(chanCall)
	}
	if info != nil {
		mqClient, err := baserpc.NewMqClient(m.App, info)
		if err != nil {
			fmt.Printf("error in moduleSession.CreateClient, %s", err.Error())
			return err
		}
		m.rpcClient.AttachMqClient(mqClient)
	}
	return err
}

func (m *moduleSession) GetId () string {
	return m.Id
}


func (m *moduleSession) GetType () string {
	return m.Type
}

func (m *moduleSession) GetClient () rpc.Client {
	return m.rpcClient
}


func (m *moduleSession) Call (id string, timeout int, args ...interface{} ) (*rpc.RetInfo, error) {
	return m.rpcClient.Call(id, timeout, args...)
}

func (m *moduleSession) AsynCall (id string, args ...interface{} ) (chan *rpc.RetInfo, error) {
	return m.rpcClient.AsynCall(id, args...)
}

