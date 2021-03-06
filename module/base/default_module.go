package basemodule

import "sync"
import (
	module "github.com/dming/lodos/module"
	"github.com/dming/lodos/conf"
	"runtime"
	log "github.com/dming/lodos/log"
)

type DefaultModule struct {
	Mi    module.Module
	CloseSig chan bool
	Wg       sync.WaitGroup
	Settings 	*conf.ModuleSettings
}

func Run(m *DefaultModule) {
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

func Destroy(m *DefaultModule) {
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






