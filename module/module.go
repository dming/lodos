package module

import "github.com/dming/lodos/rpc"
import (
	"github.com/dming/lodos/conf"
)

type Module interface {
	Version() string //模块版本
	GetType() string //模块类型
	//OnConfChanged(settings *conf.ModuleSettings)	//为以后动态服务发现做准备
	OnInit(app AppInterface, settings conf.ModuleSettings)
	OnDestroy()
	Run(closeSig chan bool)
}

type Skeleton interface {
	Init(app AppInterface, settings conf.ModuleSettings)
	GetApp() AppInterface
	GetServer() RpcServerModule
	GetServerId() string
	OnConfChanged(settings *conf.ModuleSettings)
	Call(moduleType string, funcId string, args ...interface{}) (*rpc.RetInfo , error)
	AsynCall(moduleType string, funcId string, args ...interface{}) (chan *rpc.RetInfo, error)
	Destroy()
}

type AppInterface interface {
	AddRPCSerialize(name string, rs RPCSerialize) error
	AsynCall(moduleType string, hash string, fnId string, args ...interface{}) (chan *rpc.RetInfo,  error)
	AttachChanCallToClient(id string, chanCall chan *rpc.CallInfo) error
	Call(moduleType string, hash string, fnId string, args ...interface{}) (*rpc.RetInfo,  error)
	ConfigSettings(settings conf.Config)
	Destroy() error
	GetModuleSession (filter string, hash string) (ModuleSession, error)
	GetRoute (moduleType string) (func(app AppInterface, moduleType string, hash string) (ModuleSession, error), error)
	GetRPCSerialize() (map[string]RPCSerialize)
	GetServerById(moduleId string) (ModuleSession, error)
	GetServersByType(moduleType string) []ModuleSession
	OnInit(settings conf.Config) error
	SetRoute (moduleType string, fn func(app AppInterface, moduleType string, hash string) (ModuleSession, error)) (error)
	Run(debug bool, mods ...Module) error
}

type ModuleManager interface {
	Init(app AppInterface, processId string)
	CheckModuleSettings()
	Register(mi Module)
	RegisterRunMod(mi Module)
	Destroy()
}

type ModuleSession interface {
	CreateClient(chanCall chan *rpc.CallInfo, info *conf.Rabbitmq) error
	GetId () string
	GetType () string
	GetClient () rpc.Client
	Call (id string, timeout int, args ...interface{} ) (*rpc.RetInfo, error)
	AsynCall (id string, args ...interface{} ) (chan *rpc.RetInfo, error)
}

type RpcServerModule interface {
	OnInit (app AppInterface, settings conf.ModuleSettings)
	Register(id string, f interface{})
	RegisterGo(id string, f interface{})
	GetId() string
	GetRpcServer() rpc.Server
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