package lodos

import "github.com/dming/lodos/module"
import "github.com/dming/lodos/app"

func CreateApp(version string) module.AppInterface {
	return app.NewApp(version)
}