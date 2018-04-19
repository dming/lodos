// app 表示在每个machine上运行的所有module的集合，也就是说每个machine运行一个app

package app

import (
	"github.com/dming/lodos/conf"
	"fmt"
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/module"
	"os/exec"
	"os"
	"path/filepath"
	"flag"
	 "github.com/dming/lodos/log"
	"os/signal"
	"strings"
	"math"
	"hash/crc32"
	"github.com/dming/lodos/module/base"
	"github.com/opentracing/opentracing-go"
	"github.com/dming/lodos/gate"
	"github.com/dming/lodos/rpc/base"
	"encoding/json"
	"github.com/dming/lodos/rpc/pb"
)

type resultsInfo struct {
	ErrStr string
	Results []interface{}
}

type protocolMarshalImp struct {
	data []byte
}

func (this *protocolMarshalImp) GetData() []byte {
	return this.data
}

func NewApp(version string) module.AppInterface  {
	a := new(app)
	a.routes = map[string]func(app module.AppInterface, mType string, hash string) (module.ModuleSession, error){}
	a.serverList = map[string]module.ModuleSession{}
	a.defaultRoute = func(app module.AppInterface, mType string, hash string) (module.ModuleSession, error) {
		//默认使用第一个Server
		servers := a.GetServersByType(mType)
		if len(servers) == 0 {
			return nil, fmt.Errorf("has no servers of %s", mType)
		}
		index := int(math.Abs(float64(crc32.ChecksumIEEE([]byte(hash))))) % len(servers)
		return servers[index], nil
	}
	a.rpcSerializes=map[string]module.RPCSerialize{}
	a.version = version
	return a
}

type app struct {
	//module.AppInterface
	name string
	version string
	processId string
	settings conf.Config
	manager module.ModuleManager

	serverList map[string]module.ModuleSession
	routes map[string]func(app module.AppInterface, mType string, hash string) (module.ModuleSession, error)
	defaultRoute func(app module.AppInterface, mType string, hash string) (module.ModuleSession, error)

	rpcSerializes map[string]module.RPCSerialize
	//
	getTracer           func() opentracing.Tracer
	configurationLoaded func(app module.AppInterface)
	startUp             func(app module.AppInterface)
	moduleInited        func(app module.AppInterface, module module.Module)
	judgeGuest          func(session gate.Session) bool
	protocolMarshal     func(results []interface{}, errStr string) (module.ProtocolMarshal, error)

}

func (a *app) OnInit(settings conf.Config) error {
	a.serverList = make(map[string]module.ModuleSession)
	for Type, ModuleInfos := range settings.Modules {
		for _, mInfo := range ModuleInfos {
			if m, ok := a.serverList[mInfo.Id]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("ServerId (%s) Type (%s) of the modules already exist Can not be reused ServerId (%s) Type (%s)",
					m.GetId(), m.GetType(), mInfo.Id, Type))
			}
			client, err := baserpc.NewRPCClient(a, mInfo.Id)
			if err != nil {
				continue
			}
			if a.GetProcessID() != mInfo.ProcessID {
				//同一个ProcessID下的模块直接通过local channel通信就可以了
				if mInfo.Rabbitmq != nil {
					//如果远程的rpc存在则创建一个对应的客户端
					client.NewRabbitmqRpcClient(mInfo.Rabbitmq)
				}
				if mInfo.Redis != nil && mInfo.Redis.RPCUri != "" {
					//如果远程的rpc存在则创建一个对应的客户端
					client.NewRedisRpcClient(mInfo.Redis)
				}
			}
			session := basemodule.NewModuleSession(a, mInfo.Id, Type, client)
			a.serverList[mInfo.Id] = session
			log.Info("rpcClient create success type(%s) id(%s)", Type, mInfo.Id)
		}
	}
	return nil
}

func (a *app) Run(debug bool, mods ...module.Module) error {
	wdPath := flag.String("wd", "", "Server work directory")
	confPath := flag.String("conf", "", "Server configuration file path")
	ProcessID := flag.String("pid", "development", "Server ProcessID?")
	Logdir := flag.String("log", "", "Log file directory?")
	flag.Parse() //解析输入的参数
	a.processId = *ProcessID
	ApplicationDir := ""
	if *wdPath != "" {
		_, err := os.Open(*wdPath)
		if err != nil {
			panic(err)
		}
		os.Chdir(*wdPath)
		ApplicationDir, err = os.Getwd()
	} else {
		var err error
		ApplicationDir, err = os.Getwd()
		if err != nil {
			file, _ := exec.LookPath(os.Args[0])
			ApplicationPath, _ := filepath.Abs(file)
			ApplicationDir, _ = filepath.Split(ApplicationPath)
		}

	}// get ApplicationDir

	defaultConfPath := fmt.Sprintf("%s/conf/server.conf", ApplicationDir)
	defaultLogPath := fmt.Sprintf("%s/logs", ApplicationDir)

	if *confPath == "" {
		*confPath = defaultConfPath
	}

	if *Logdir == "" {
		*Logdir = defaultLogPath
	}

	f, err := os.Open(*confPath)
	if err != nil {
		panic(err)
	}

	_, err = os.Open(*Logdir)
	if err != nil {
		//文件不存在
		err := os.Mkdir(*Logdir, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("Server configuration file path :", *confPath)
	conf.LoadConfig(f.Name()) //加载配置文件
	a.ConfigSettings(conf.Conf)  //配置信息
	log.InitBeego(debug, *ProcessID, *Logdir, conf.Conf.Log)

	log.Info("Lodos %v starting up", a.version)

	if a.configurationLoaded != nil {
		a.configurationLoaded(a)
	}

	manager := NewModuleManager()
	//manager.RegisterRunMod(modules.TimerModule()) //注册时间轮模块 每一个进程都默认运行
	// module
	for i := 0; i < len(mods); i++ {
		mods[i].OnAppConfigurationLoaded(a)
		manager.Register(mods[i])
	}
	a.OnInit(a.settings)
	manager.Init(a, *ProcessID)
	if a.startUp != nil {
		a.startUp(a)
	}
	if a.judgeGuest == nil {
		log.Warning("App.judgeGuest is still nil, it would cause all user just be guest ")
	}
	log.Info("Server started : %s", *confPath)
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	manager.Destroy()
	a.OnDestroy()
	log.Info("Lodos closing down (signal: %v)", sig)
	return nil
}

func (a *app) OnDestroy() error {
	for id, session := range a.serverList {
		err := session.GetClient().Done()
		if err != nil {
			log.Warning("rpcClient close fail type(%s) id(%s)", session.GetType(), id)
		} else {
			log.Info("rpcClient close success type(%s) id(%s)", session.GetType(), id)
		}
	}
	return nil
}


func (a *app) ConfigSettings(settings conf.Config) {
	a.settings = settings
}
func (a *app) GetSettings() conf.Config {
	return a.settings
}
func (a *app) GetProcessID() string {
	return a.processId
}

func (a *app) RegisterLocalClient(serverId string, server rpc.RPCServer) error {
	if session, ok := a.serverList[serverId]; ok {
		err := session.GetClient().NewLocalRpcClient(server)
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("Server(%s) Not Found", serverId)
	}
}

//  Route(moduleType string, fn func(app AppInterface, Type string, hash string) ModuleSession) error
func (a *app) SetRoute (moduleType string, fn func(app module.AppInterface, moduleType string, hash string) (module.ModuleSession, error)) {
	a.routes[moduleType] = fn
}
func (a *app) GetRoute (moduleType string) (func(app module.AppInterface, moduleType string, hash string) (module.ModuleSession, error)) {
	if _, ok := a.routes[moduleType]; ok {
		return a.routes[moduleType]
	} else {
		return a.defaultRoute
	}
}
func (a *app) GetServerById(moduleId string) (module.ModuleSession, error) {
	if ms, ok := a.serverList[moduleId]; ok {
		return ms, nil
	} else {
		return nil, fmt.Errorf("%s not found.", moduleId)
	}
}
/**
filter		 调用者服务类型    moduleType|moduleType@moduleID
Type	   	想要调用的服务类型
*/
func (a *app) GetServersByType(moduleType string) []module.ModuleSession {
	sessions := make([]module.ModuleSession, 0)
	for _, session := range a.serverList {
		if session.GetType() == moduleType {
			sessions = append(sessions, session)
		}
	}
	return sessions
}
//GetModuleSession (filter string, hash string) (ModuleSession, error)
func (a *app) GetRouteServer(filter string, hash string) (module.ModuleSession, error) {
	sl := strings.Split(filter, "@")
	if len(sl) == 2 {
		moduleID := sl[1]
		if moduleID != "" {
			return a.GetServerById(moduleID)
		}
	}
	moduleType := sl[0]
	route := a.GetRoute(moduleType) //route is a function
	s, _ := route(a, moduleType, hash)  // s is module.ServerSession
	if s == nil {
		return nil, fmt.Errorf("Server(type : %s) Not Found", moduleType)
	}
	return s, nil
} //获取经过筛选过的服务


func (a *app) AddRPCSerialize(name string, rs module.RPCSerialize) error {
	if _, ok := a.rpcSerializes[name]; ok{
		return fmt.Errorf("The name(%s) has been occupied",name)
	}
	a.rpcSerializes[name] = rs
	return nil
}
func (a *app) GetRPCSerialize() (map[string]module.RPCSerialize) {
	return a.rpcSerializes
}

func (a *app) DefaultTracer(fn func() opentracing.Tracer) {
	a.getTracer = fn
}
func (a *app) GetTracer() opentracing.Tracer {
	if a.getTracer != nil {
		return a.getTracer()
	}
	return nil
}

func (a *app) SetJudgeGuest(judgeGuest func(session gate.Session) bool) {
	a.judgeGuest = judgeGuest
}
func (a *app) GetJudgeGuest() func(session gate.Session) bool {
	return a.judgeGuest
}

func (a *app) SetProtocolMarshal(protocolMarshal func(Results []interface{}, errStr string) (module.ProtocolMarshal, error)) {
	a.protocolMarshal = protocolMarshal
}
func (a *app) NewProtocolMarshal(data []byte) module.ProtocolMarshal {
	return &protocolMarshalImp{
		data: data,
	}
}
func (a *app) ProtocolMarshal(results []interface{}, errStr string) (module.ProtocolMarshal, error) {
	if a.protocolMarshal != nil {
		return a.protocolMarshal(results, errStr)
	}

	r := &resultsInfo{
		ErrStr: errStr,
		Results: results,
	}
	body, err := json.Marshal(r)
	if err != nil {
		return nil, err
	} else {
		return a.NewProtocolMarshal(body), nil
	}
}


func (a *app) GetModuleInited() func(app module.AppInterface, module module.Module) {
	return a.moduleInited
}
func (a *app) SetModuleInited(fn func(app module.AppInterface, module module.Module)) {
	a.moduleInited = fn
}

func (a *app) SetConfigurationLoaded(fn func(app module.AppInterface)) {
	a.configurationLoaded = fn
}
func (a *app) SetStartup(fn func(app module.AppInterface)) {
	a.startUp = fn
}


func (a *app) RpcCall(m module.FullModule, moduleType string, _func string, params ...interface{}) ([]interface{}, error) {
	server, err := a.GetRouteServer(moduleType, m.GetServerId())
	if err != nil {
		return nil, err
	}
	return server.Call(_func, params...)
}


func (a *app) RpcSyncCall(m module.FullModule, moduleType string, _func string, params ...interface{}) (chan rpcpb.ResultInfo, error) {
	server, err := a.GetRouteServer(moduleType, m.GetServerId())
	if err != nil {
		return nil, err
	}
	return server.SyncCall(_func, params...)
}

func (a *app) RpcCallNR(m module.FullModule, moduleType string, _func string, params ...interface{}) (err error) {
	server, err := a.GetRouteServer(moduleType, m.GetServerId())
	if err != nil {
		return err
	}
	return server.CallNR(_func, params...)
}

func (a *app) RpcCallArgs(m module.FullModule, moduleType string, _func string, ArgsType []string, Args [][]byte) ([]interface{}, error) {
	server, err := a.GetRouteServer(moduleType, m.GetServerId())
	if err != nil {
		return nil, err
	}
	return server.CallArgs(_func, ArgsType, Args)
}

func (a *app) RpcSyncCallArgs(m module.FullModule, moduleType string, _func string, ArgsType []string, Args [][]byte) (chan rpcpb.ResultInfo, error) {
	server, err := a.GetRouteServer(moduleType, m.GetServerId())
	if err != nil {
		return nil, err
	}
	return server.SyncCallArgs(_func, ArgsType, Args)
}

func (a *app) RpcCallArgsNR(m module.FullModule, moduleType string, _func string, ArgsType []string, Args [][]byte) (err error) {
	server, err := a.GetRouteServer(moduleType, m.GetServerId())
	if err != nil {
		return err
	}
	return server.CallArgsNR(_func, ArgsType, Args)
}






/*
func (a *app) GetModuleSession (filter string, hash string) (module.ModuleSession, error) {
	sl := strings.Split(filter, "@")
	if len(sl) == 2 {
		moduleId := sl[1]
		if moduleId != "" {
			return a.GetServerById(moduleId)
		}
	}
	moduleType := sl[0]
	route, _ := a.GetRoute(moduleType)
	ms, err := route(a, moduleType, hash)
	if err != nil {
		return nil, err
	} else {
		return ms, nil
	}

}

func (a *app) Call(moduleType string, hash string, fnId string, args ...interface{}) (*rpc.RetInfo,  error) {
	ms, err := a.GetModuleSession(moduleType, hash)
	if err != nil {
		return nil, err
	}
	return ms.GetClient().Call(fnId, 1,  args...)
}

func (a *app) AsynCall(moduleType string, hash string, fnId string, args ...interface{}) (chan *rpc.RetInfo,  error) {
	ms, err := a.GetModuleSession(moduleType, hash)
	if err != nil {
		return nil, err
	}
	return ms.GetClient().AsynCall(fnId, args...)
}
*/