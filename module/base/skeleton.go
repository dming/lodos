package basemodule

import (
	"sync"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/module"
)

func NewSkeleton() *Skeleton {
	s := new(Skeleton)
	return s
}

type Skeleton struct {
	//module.Skeleton
	App module.AppInterface
	settings conf.ModuleSettings
	hash string
	server module.RpcServerModule
	rwmutex *sync.RWMutex
}

func (s *Skeleton) OnInit(app module.AppInterface, settings conf.ModuleSettings)  {
	s.App = app
	s.settings = settings
	s.hash = settings.Id
	s.GetServer().OnInit(app, settings)
}

func (s *Skeleton) Destroy()  {
	s.GetServer().OnDestroy()
}

func (s *Skeleton)GetServer() module.RpcServerModule  {
	if s.server == nil {
		return new(rpcServerModule)
	}
	return s.server
}

func (s *Skeleton) GetServerId() string {
	//很关键,需要与配置文件中的Module配置对应
	return s.settings.Id
}

func (s *Skeleton) GetApp() module.AppInterface {
	return s.App
}

func (s *Skeleton)OnConfChanged(settings *conf.ModuleSettings)  {

}

func (s *Skeleton) Call(moduleType string, funcId string, args ...interface{}) (*rpc.RetInfo , error) {
	ms, err := s.App.GetModuleSession(moduleType, s.hash)
	if err != nil {
		return nil, err
	}
	return ms.GetClient().Call(funcId, 1, args...)
}

func (s *Skeleton) AsynCall(moduleType string, funcId string, args ...interface{}) (chan *rpc.RetInfo, error) {
	client, err := s.App.GetModuleSession(moduleType, s.hash)
	if err != nil {
		return nil, err
	}
	return client.GetClient().AsynCall(funcId, args...)
}