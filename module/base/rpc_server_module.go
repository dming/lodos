package basemodule

import (
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/rpc/base"
	"github.com/dming/lodos/log"
)

func NewRpcServerModule(app module.AppInterface, module module.Module, settings *conf.ModuleSettings) module.RpcServerModule {
	rs := new(rpcServerModule)
	rs.OnInit(app, module, settings)
	return  rs
}


type rpcServerModule struct {
	App module.AppInterface
	server rpc.RPCServer
	settings *conf.ModuleSettings
}

func (rs *rpcServerModule) OnInit (app module.AppInterface, module module.Module, settings *conf.ModuleSettings)  {
	rs.settings = settings
	//默认会创建一个本地的RPC
	server, err := baserpc.NewRPCServer(app, module)
	if err != nil {
		log.Error("fail in dialing rpc server : %s", err)
	}
	logInfo := "local, "
	if settings.Rabbitmq != nil {
		//存在远程rpc的配置
		server.NewRabbitmqRpcServer(settings.Rabbitmq)
		logInfo += "rabbitmq, "
	}
	if settings.Redis != nil && settings.Redis.RPCUri != ""  {
		//存在远程redis rpc的配置
		server.NewRedisRpcServer(settings.Redis)
		logInfo += "redis"
	}

	rs.server = server
	//向App注册该RpcServerModule
	//err = app.AttachChanCallToClient(settings.id, rs.server.GetChanCall())
	err = app.RegisterLocalClient(settings.Id, server)
	if err != nil {
		log.Warning("RegisterLocalClient: id(%s) error(%s)", settings.Id, err)
	}
	log.Info("RPCServer init success id(%s), include %s", rs.settings.Id, logInfo)
}

func (rs *rpcServerModule) OnDestroy() {
	if rs.server != nil {
		err := rs.server.Done()
		if err != nil {
			log.Warning("RPCServer close fail id(%s) error(%s)", rs.settings.Id, err)
		} else {
			log.Info("RPCServer close success id(%s)", rs.settings.Id)
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


func (rs *rpcServerModule) GetRpcServer() rpc.RPCServer {
	return rs.server
}