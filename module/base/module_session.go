package basemodule

import (
	"rpc"
	"fmt"
	"conf"
	"module"
	"rpc/base"
)

type moduleSession struct {
	Id string
	App module.AppInterface
	Type string
	RpcClient rpc.Client
}

func NewModuleSession(app module.AppInterface, id string, typ string, chanCall chan *rpc.CallInfo, info *conf.Rabbitmq) (module.ModuleSession, error) {
	m := new(moduleSession)
	m.App = app
	m.Id = id
	m.Type = typ
	err := m.CreateClient(chanCall, info)
	return m, err
}


func (m *moduleSession) CreateClient(chanCall chan *rpc.CallInfo, info *conf.Rabbitmq) error {
	var err error
	m.RpcClient, err = baserpc.NewClient()
	if err != nil {
		fmt.Printf("error in moduleSession.CreateClient, %s", err.Error())
		return err
	}
	if chanCall != nil {
		m.RpcClient.AttachChanCall(chanCall)
	}
	if info != nil {
		mqClient, err := baserpc.NewMqClient(m.App, info)
		if err != nil {
			fmt.Printf("error in moduleSession.CreateClient, %s", err.Error())
			return err
		}
		m.RpcClient.AttachMqClient(mqClient)
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
	return m.RpcClient
}

