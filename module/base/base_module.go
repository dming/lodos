package basemodule

import "sync"
import (
	module "module"
	"conf"
	"runtime"
	log "mlog"
)

type BaseModule struct {
	Mi    module.Module
	CloseSig chan bool
	Wg       sync.WaitGroup
	Settings 	*conf.ModuleSettings
}

func Run(m *BaseModule) {
	defer func() {
		if r := recover(); r != nil {
			if false {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()

	m.Mi.Run(m.CloseSig)
	m.Wg.Done()
}

func Destroy(m *BaseModule) {
	defer func() {
		if r := recover(); r != nil {
			if false {
				//
			} else {
				//
			}
		}
	}()
	m.Mi.OnDestroy()
}






