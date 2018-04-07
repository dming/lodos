package basemodule

import (
	"encoding/json"
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/conf"
	"sync"
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/rpc/pb"
	"time"
)

type StatisticalMethod struct {
	Name        string //方法名
	StartTime   int64  //开始时间
	EndTime     int64  //结束时间
	MinExecTime int64  //最短执行时间
	MaxExecTime int64  //最长执行时间
	ExecTotal   int    //执行总次数
	ExecTimeout int    //执行超时次数
	ExecSuccess int    //执行成功次数
	ExecFailure int    //执行错误次数
}

func LoadStatisticalMethod(j string) map[string]*StatisticalMethod {
	sm := map[string]*StatisticalMethod{}
	err := json.Unmarshal([]byte(j), &sm)
	if err == nil {
		return sm
	} else {
		return nil
	}
}
/**
	Implement FullModule except Module
 */
type BaseModule struct {
	//module.Module
	app      module.AppInterface
	settings *conf.ModuleSettings
	hash     string //server id
	server   module.RpcServerModule
	rwmutex  *sync.RWMutex
	//useless now
	listener rpc.RPCListener
	statistical map[string]*StatisticalMethod //统计
}

func (m *BaseModule) OnInit(app module.AppInterface, module module.Module, settings *conf.ModuleSettings) {
	m.app = app
	m.settings = settings
	m.statistical = map[string]*StatisticalMethod{}
	m.GetServer().OnInit(app, module, settings)
	m.GetServer().GetRpcServer().SetListener(module)
}

func (m *BaseModule) OnDestroy() {
	//注销模块
	//一定别忘了关闭RPC
	m.GetServer().OnDestroy()
}

func (m *BaseModule) GetApp() module.AppInterface {
	return m.app
}

func (m *BaseModule) GetServer() module.RpcServerModule {
	return m.server
}

func (m *BaseModule) GetServerId() string {
	return m.server.GetId()
}

func (m *BaseModule) GetModuleSetting() *conf.ModuleSettings {
	return m.settings
}

func (m *BaseModule) RpcCall(moduleType string, _func string, params ...interface{}) ([]interface{}, error) {
	server, err := m.app.GetRouteServer(moduleType, m.hash)
	if err != nil {
		return nil, err
	}
	return server.Call(_func, params...)
}


func (m *BaseModule) RpcSyncCall(moduleType string, _func string, params ...interface{}) (chan rpcpb.ResultInfo, error) {
	server, err := m.app.GetRouteServer(moduleType, m.hash)
	if err != nil {
		return nil, err
	}
	return server.SyncCall(_func, params...)
}

func (m *BaseModule) RpcCallNR(moduleType string, _func string, params ...interface{}) (err error) {
	server, err := m.app.GetRouteServer(moduleType, m.hash)
	if err != nil {
		return err
	}
	return server.CallNR(_func, params...)
}

func (m *BaseModule) RpcCallArgs(moduleType string, _func string, ArgsType []string, Args [][]byte) ([]interface{}, error) {
	server, err := m.app.GetRouteServer(moduleType, m.hash)
	if err != nil {
		return nil, err
	}
	return server.CallArgs(_func, ArgsType, Args)
}

func (m *BaseModule) RpcSyncCallArgs(moduleType string, _func string, ArgsType []string, Args [][]byte) (chan rpcpb.ResultInfo, error) {
	server, err := m.app.GetRouteServer(moduleType, m.hash)
	if err != nil {
		return nil, err
	}
	return server.SyncCallArgs(_func, ArgsType, Args)
}

func (m *BaseModule) RpcCallArgsNR(moduleType string, _func string, ArgsType []string, Args [][]byte) (err error) {
	server, err := m.app.GetRouteServer(moduleType, m.hash)
	if err != nil {
		return err
	}
	return server.CallArgsNR(_func, ArgsType, Args)
}

func (m *BaseModule) GetRouteServer(filter string, hash string) (module.ModuleSession, error) {
	//return m.app.GetRouteServer(moduleType, hash)
	return nil, nil
}

func (m *BaseModule) GetStatistical() (statistical string, err error) {
	m.rwmutex.Lock()
	//重置
	now := time.Now().UnixNano()
	for _, s := range m.statistical {
		s.EndTime = now
	}
	b, err := json.Marshal(m.statistical)
	if err == nil {
		statistical = string(b)
	}

	//重置
	//for _,s:=range m.statistical{
	//	s.StartTime=now
	//	s.ExecFailure=0
	//	s.ExecSuccess=0
	//	s.ExecTimeout=0
	//	s.ExecTotal=0
	//	s.MaxExecTime=0
	//	s.MinExecTime=math.MaxInt64
	//}
	m.rwmutex.Unlock()
	return
}

func (m *BaseModule) GetExecuting() int64 {
	return m.GetServer().GetRpcServer().GetExecuting()
}

func (m *BaseModule) OnAppConfigurationLoaded(app module.AppInterface) {
	//
}

func (m *BaseModule) OnConfChanged(settings *conf.ModuleSettings) {
	//
}