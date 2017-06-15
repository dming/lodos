package app

import (
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/module/base"
	"github.com/dming/lodos/conf"
	"fmt"
)

func NewModuleManager() module.ModuleManager {
	m := new(moduleManager)
	return m
}

type moduleManager struct {
	app module.AppInterface
	mods []*basemodule.BaseModule
	runMods []*basemodule.BaseModule
}

func (mg *moduleManager) Init(app module.AppInterface, processId string) {
	//log.Info("This service ProcessID is [%s]", ProcessID)
	mg.app = app
	mg.CheckModuleSettings() //配置文件规则检查

	for i := 0; i < len(mg.mods); i++ {
		for Type, modSettings := range conf.Conf.Modules {
			if mg.mods[i].Mi.GetType() == Type {
				//匹配
				for _, setting := range modSettings {
					//这里可能有BUG 公网IP和局域网IP处理方式可能不一样,先不管
					if processId == setting.ProcessID {
						mg.runMods = append(mg.runMods, mg.mods[i]) //这里加入能够运行的组件
						mg.mods[i].Settings = *setting
					}
				}
				break //跳出内部循环
			}
		}
	}

	for i := 0; i < len(mg.runMods); i++ {
		m := mg.runMods[i]
		m.Mi.OnInit(app, m.Settings)
		m.Wg.Add(1)
		go basemodule.Run(m)
	}
	//timer.SetTimer(3, mer.ReportStatistics, nil) //统计汇报定时任务

}


func (mg *moduleManager) Register(mi module.Module) {
	m := new(basemodule.BaseModule)
	m.Mi = mi
	m.CloseSig = make(chan bool, 1)

	mg.mods = append(mg.mods, m)
}

func (mg *moduleManager) RegisterRunMod(mi module.Module) {
	m := new(basemodule.BaseModule)
	m.Mi = mi
	m.CloseSig = make(chan bool, 1)

	mg.runMods = append(mg.runMods, m)
}


/**
module配置文件规则检查
1. ID全局必须唯一
2. 每一个类型的Module列表中ProcessID不能重复
*/
func (mg *moduleManager) CheckModuleSettings() {
	gid := map[string]string{} //用来保存全局ID-ModuleType
	for Type, modSettings := range conf.Conf.Modules {
		pid := map[string]string{} //用来保存模块中的 ProcessID-ID
		for _, setting := range modSettings {
			if Stype, ok := gid[setting.Id]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("ID (%s) been used in modules of type [%s] and cannot be reused", setting.Id, Stype))
			} else {
				gid[setting.Id] = Type
			}

			if Id, ok := pid[setting.ProcessID]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("In the list of modules of type [%s], ProcessID (%s) has been used for ID module for (%s)", Type, setting.ProcessID, Id))
			} else {
				pid[setting.ProcessID] = setting.Id
			}
		}
	}
}

func (mg *moduleManager) Destroy() {
	for i := len(mg.runMods) - 1; i >= 0; i-- {
		m := mg.runMods[i]
		m.CloseSig <- true
		m.Wg.Wait()
		basemodule.Destroy(m)
	}
}

