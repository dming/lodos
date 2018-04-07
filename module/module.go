package module

import "github.com/dming/lodos/rpc"
import (
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/rpc/pb"
	"github.com/opentracing/opentracing-go"
	"github.com/dming/lodos/gate"
)

type ProtocolMarshal interface {
	GetData() []byte
}

type Module interface {
	//GetApp() AppInterface
	Version() string //模块版本
	GetType() string //模块类型
	OnAppConfigurationLoaded(app AppInterface)
	OnConfChanged(settings *conf.ModuleSettings)	//为以后动态服务发现做准备
	OnInit(app AppInterface, settings *conf.ModuleSettings)
	OnDestroy()
	Run(closeSig chan bool)
}

type FullModule interface {
	Module

	GetApp() AppInterface
	GetServer() RpcServerModule
	GetServerId() string
	GetModuleSetting() (*conf.ModuleSettings)
	// rpc invoke
	RpcCall(moduleType string, _func string, params ...interface{}) ([]interface{}, error)
	RpcSyncCall(moduleType string, _func string, params ...interface{}) (chan rpcpb.ResultInfo, error)
	RpcCallNR(moduleType string, _func string, params ...interface{}) (err error)
	RpcCallArgs(moduleType string, _func string, ArgsType []string, Args [][]byte) ([]interface{}, error)
	RpcSyncCallArgs(moduleType string, _func string, ArgsType []string, Args [][]byte) (chan rpcpb.ResultInfo, error)
	RpcCallArgsNR(moduleType string, _func string, ArgsType []string, Args [][]byte) (err error)
	/**
		filter		 调用者服务类型    moduleType|moduleType@moduleID
		Type	   	想要调用的服务类型
	**/
	GetRouteServer(filter string, hash string) (ModuleSession, error)
	GetStatistical() (statistical string, err error)
	GetExecuting() int64
}

/*
//Module 应该包含skeleton，
//设立Skelton是为了在server的Module里直接继承struct Skeleton，方便开发
type Skeleton interface {
	Init(app AppInterface, settings conf.ModuleSettings)
	GetApp() AppInterface
	GetServer() RpcServerModule
	GetServerId() string
	OnConfChanged(settings *conf.ModuleSettings)
	Call(moduleType string, funcId string, args ...interface{}) (*rpc.RetInfo , error)
	AsynCall(moduleType string, funcId string, args ...interface{}) (chan *rpc.RetInfo, error)
	//Destroy()

	////

}*/

type AppInterface interface {
	ConfigSettings(settings conf.Config)
	GetSettings() conf.Config
	GetProcessID() string


	OnInit(settings conf.Config) error
	Run(debug bool, mods ...Module) error
	OnDestroy() error

	RegisterLocalClient(serverId string, server rpc.RPCServer) error
	//  Route(moduleType string, fn func(app AppInterface, Type string, hash string) ModuleSession) error
	SetRoute (moduleType string, fn func(app AppInterface, moduleType string, hash string) (ModuleSession, error))
	GetRoute (moduleType string) (func(app AppInterface, moduleType string, hash string) (ModuleSession, error))
	GetServerById(moduleId string) (ModuleSession, error)
	/**
	filter		 调用者服务类型    moduleType|moduleType@moduleID
	Type	   	想要调用的服务类型
	*/
	GetServersByType(moduleType string) []ModuleSession
	//GetModuleSession (filter string, hash string) (ModuleSession, error)
	GetRouteServer(filter string, hash string) (ModuleSession, error) //获取经过筛选过的服务


	// rpc invoke
	RpcCall(module FullModule, moduleType string, _func string, params ...interface{}) ([]interface{}, error)
	RpcSyncCall(module FullModule, moduleType string, _func string, params ...interface{}) (chan rpcpb.ResultInfo, error)
	RpcCallNR(module FullModule, moduleType string, _func string, params ...interface{}) (err error)
	RpcCallArgs(module FullModule, moduleType string, _func string, ArgsType []string, Args [][]byte) ([]interface{}, error)
	RpcSyncCallArgs(module FullModule, moduleType string, _func string, ArgsType []string, Args [][]byte) (chan rpcpb.ResultInfo, error)
	RpcCallArgsNR(module FullModule, moduleType string, _func string, ArgsType []string, Args [][]byte) (err error)


	AddRPCSerialize(name string, rs RPCSerialize) error
	GetRPCSerialize() (map[string]RPCSerialize)

	DefaultTracer(func() opentracing.Tracer)
	GetTracer() opentracing.Tracer

	SetJudgeGuest(judgeGuest func(session gate.Session) bool)
	GetJudgeGuest() func(session gate.Session) bool

	SetProtocolMarshal(protocolMarshal func(Result []interface{}, Error string) (ProtocolMarshal, error))
	NewProtocolMarshal(data []byte) ProtocolMarshal
	ProtocolMarshal(results []interface{}, errStr string) (ProtocolMarshal, error)

	SetModuleInited(fn func(app AppInterface, module Module))
	GetModuleInited() func(app AppInterface, module Module)
	SetConfigurationLoaded(fn func(app AppInterface))
	SetStartup(fn func(app AppInterface))

}

type ModuleManager interface {
	Init(app AppInterface, processId string)
	CheckModuleSettings()
	Register(mi Module)
	RegisterRunMod(mi Module)
	Destroy()
}

type ModuleSession interface {
	GetApp () AppInterface
	GetId () string
	GetType () string
	GetClient () rpc.RPCClient
	// rpc invoke
	Call(_func string, params ...interface{}) ([]interface{}, error)
	SyncCall(_func string, params ...interface{}) (chan rpcpb.ResultInfo, error)
	CallNR(_func string, params ...interface{}) (err error)
	CallArgs(_func string, ArgsType []string, Args [][]byte) ([]interface{}, error)
	SyncCallArgs(_func string, ArgsType []string, Args [][]byte) (chan rpcpb.ResultInfo, error)
	CallArgsNR(_func string, ArgsType []string, Args [][]byte) (err error)
}

type RpcServerModule interface {
	OnInit (app AppInterface, module Module, settings *conf.ModuleSettings)
	Register(id string, f interface{})
	RegisterGo(id string, f interface{})
	GetId() string
	GetRpcServer() rpc.RPCServer
	OnDestroy()
}


/**
rpc 自定义参数序列化接口
 */
type RPCSerialize interface {
	/**
	序列化 结构体-->[]byte
	param 需要序列化的参数值
	@return ptype 当能够序列化这个值,并且正确解析为[]byte时 返回改值正确的类型,否则返回 ""即可
	@return p 解析成功得到的数据, 如果无法解析该类型,或者解析失败 返回nil即可
	@return err 无法解析该类型,或者解析失败 返回错误信息
	 */
	Serialize(param interface{}) (ptype string, p []byte, err error)
	/**
	反序列化 []byte-->结构体
	ptype 参数类型 与Serialize函数中ptype 对应
	b   参数的字节流
	@return param 解析成功得到的数据结构
	@return err 无法解析该类型,或者解析失败 返回错误信息
	 */
	Deserialize(ptype string,b []byte)(param interface{},err error)
	/**
	返回这个接口能够处理的所有类型
	 */
	GetTypes()([]string)
}