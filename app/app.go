// app 表示在每个machine上运行的所有module的集合，也就是说每个machine运行一个app

package app

import (
	"conf"
	"fmt"
	"rpc"
	"module"
	"os/exec"
	"os"
	"path/filepath"
	"flag"
	log "mlog"
	"os/signal"
	"strings"
	"math"
	"hash/crc32"
	"module/base"
)



func NewApp(version string) module.AppInterface  {
	a := new(app)
	a.routes = map[string]func(app module.AppInterface, mType string, hash string) (module.ModuleSession, error){}
	a.serverList = map[string]module.ModuleSession{}
	a.defaultRoute = func(app module.AppInterface, mType string, hash string) (module.ModuleSession, error) {
		//默认使用第一个Server
		servers := app.GetServersByType(mType)
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
	name string
	version string
	settings conf.Config

	serverList map[string]module.ModuleSession
	routes map[string]func(app module.AppInterface, mType string, hash string) (module.ModuleSession, error)
	defaultRoute func(app module.AppInterface, mType string, hash string) (module.ModuleSession, error)
	rpcSerializes map[string]module.RPCSerialize
	manager module.ModuleManager
}

func (a *app) OnInit(settings conf.Config) error {
	a.serverList = make(map[string]module.ModuleSession)
	for Type, ModuleInfos := range settings.Modules {
		for _, module := range ModuleInfos {
			if m, ok := a.serverList[module.Id]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("ServerId (%s) Type (%s) of the modules already exist Can not be reused ServerId (%s) Type (%s)",
					m.GetId(), m.GetType(), module.Id, Type))
			}

			session, err := basemodule.NewModuleSession(a, module.Id, Type, nil, module.RabbitmqInfo)
			if err != nil {
				//log err
				continue
			}
			a.serverList[module.Id] = session
			log.Info("RPCClient create success type(%s) id(%s)", Type, module.Id)
		}
	}
	return nil
}

func (a *app) Run(debug bool, mods ...module.Module) error {
	file, _ := exec.LookPath(os.Args[0])
	applicationPath, _ := filepath.Abs(file)
	applicationDir, _ := filepath.Split(applicationPath)
	defaultPath := fmt.Sprintf("%sconf/server.conf", applicationDir)
	confPath := flag.String("conf", defaultPath, "Server config file path.")
	processId := flag.String("pid", "development", "Server proccess id")
	logDir := flag.String("log", fmt.Sprintf("%slogs", applicationDir), "log file")
	flag.Parse()//parse the input args

	f, err := os.Open(*confPath)
	if err != nil {
		panic(err)
	}

	_, err = os.Open(*logDir)
	if err != nil {
		//file not exist
		err = os.Mkdir(*logDir, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}
	//log.Init(debug, *processId, *logDir)
	//log.Info("Server configuration file path [%s]", *confPath)

	conf.LoadConfig(f.Name()) //加载配置文件
	a.ConfigSettings(conf.Conf)  //配置信息

	//log.Info("the server starting...")
	a.manager = NewModuleManager()

	for _, v := range mods {
		a.manager.RegisterRunMod(v)
	}
	a.OnInit(a.settings)
	a.manager.Init(a, *processId)

	//close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	a.manager.Destroy()
	a.Destroy()
	//log.Info("server closing down (signal %v)", sig)
	fmt.Printf("server closing down (signal %v)", sig)
	return nil
}

func (a *app) Destroy() error {
	for id, session := range a.serverList {
		err := session.GetClient().Close()
		if err != nil {
			//log.Warning("RPCClient close fail type(%s) id(%s)", session.GetType(), id)
			fmt.Printf("RPCClient close fail type(%s) id(%s)", session.GetType(), id)
		} else {
			//log.Info("RPCClient close success type(%s) id(%s)", session.GetType(), id)
		}
	}
	return nil
}

func (a *app) AttachChanCallToClient(id string, chanCall chan *rpc.CallInfo) error {
	if session, ok := a.serverList[id]; ok {
		session.GetClient().AttachChanCall(chanCall)
		return nil
	} else {
		return fmt.Errorf("Server(%s) Not Found", id)
	}
}

func (a *app) ConfigSettings(settings conf.Config) {
	a.settings = settings
}

func (a *app) SetRoute (moduleType string, fn func(app module.AppInterface, moduleType string, hash string) (module.ModuleSession, error)) (error)  {
	a.routes[moduleType] = fn
	return nil
}

func (a *app) GetRoute (moduleType string) (func(app module.AppInterface, moduleType string, hash string) (module.ModuleSession, error), error)  {
	if _, ok := a.routes[moduleType]; ok {
		return a.routes[moduleType], nil
	} else {
		return a.defaultRoute, nil
	}
}

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
	ms, err := route(*a, moduleType, hash)
	if err != nil {
		return nil, err
	} else {
		return ms, nil
	}

}

func (a *app) GetServerById(moduleId string) (module.ModuleSession, error)  {
	if ms, ok := a.serverList[moduleId]; ok {
		return ms, nil
	} else {
		return nil, fmt.Errorf("%s not found.", moduleId)
	}
}

func (a *app) GetServersByType(moduleType string) []module.ModuleSession {
	sessions := make([]module.ModuleSession, 0)
	for _, session := range a.serverList {
		if session.GetType() == moduleType {
			sessions = append(sessions, session)
		}
	}
	return sessions
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