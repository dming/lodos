package lodos

import "github.com/dming/lodos/module"
import "github.com/dming/lodos/app"

func CreateApp() module.AppInterface {
	return app.NewApp("1.0.0")
}