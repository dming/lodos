package basemodule

import (
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/rpc/base"
	log "github.com/dming/lodos/mlog"
)

func NewRpcServerModule(app module.AppInterface, settings conf.ModuleSettings) module.RpcServerModule {
	rs := new(rpcServerModule)
	rs.OnInit(app, settings)
	return  rs
}

type rpcServerModule struct {
	App module.AppInterface
	server rpc.Server
	settings conf.ModuleSettings
}

func (rs *rpcServerModule) OnInit (app module.AppInterface, settings conf.ModuleSettings)  {
	rs.settings = settings
	server := baserpc.NewServer() //默认会创建一个本地的RPC
	if settings.RabbitmqInfo != nil {
		//存在远程rpc的配置
		mqServer, err := baserpc.NewMqServer(rs.App, settings.RabbitmqInfo, server.GetChanCall())
		if err != nil {
			log.Error("create mq server fail !!")
			panic("create mq server fail !!")
		}
		server.AttachMqServer(mqServer)
	}

	rs.server = server
	//向App注册该RpcServerModule
	err := app.AttachChanCallToClient(settings.Id, rs.server.GetChanCall())
	if err != nil {
		log.Warning("RegisterLocalClient: id(%s) error(%s)", settings.Id, err)
	}
	log.Info("RPCServer init success id(%s)", rs.settings.Id)
}

func (rs *rpcServerModule) OnDestroy() {
	if rs.server != nil {
		err := rs.server.Close()
		if err != nil {
			//log.Warning("RPCServer close fail id(%s) error(%s)", s.settings.Id, err)
		} else {
			//log.Info("RPCServer close success id(%s)", s.settings.Id)
		}
		rs.server = nil
	}
}

func (rs *rpcServerModule) Register(id string, f interface{}) {
	if rs.server == nil {
		panic("invalid RPCServer")
	}
	rs.server.Register(id, f)
}

func (rs *rpcServerModule) RegisterGo(id string, f interface{}) {
	if rs.server == nil {
		panic("invalid RPCServer")
	}
	rs.server.RegisterGo(id, f)
}


func (rs *rpcServerModule) GetId() string {
	return rs.settings.Id
}


func (rs *rpcServerModule) GetRpcServer() rpc.Server {
	return rs.server
}