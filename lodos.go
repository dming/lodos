package lodos

import "module"
import (
	"app"
)
func CreateApp() module.AppInterface {
	return app.NewApp("1.0.0")
}