package base

import (
	"github.com/dming/lodos/addon/ver"
	"github.com/dming/lodos/module/base"
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/conf"
)

func NewExchangeVerifyModule(params ...interface{}) *exchangeVerifyModule {
	panic("implement me")
	return &exchangeVerifyModule{}
}

type exchangeVerifyModule struct {
	//module.FullModule
	basemodule.BaseModule
	exchange_verify ver.ExchangeVerify
	db_handler ver.RedisDBHandler
}

func (this *exchangeVerifyModule) Version() string {
	return "1.0.0"
}

func (this *exchangeVerifyModule) GetType() string {
	return "ExchangeVerify"
}

func (this *exchangeVerifyModule) OnInit(app module.AppInterface, settings *conf.ModuleSettings) {
	this.BaseModule.OnInit(app, this, settings)
	this.db_handler.OnInit()
}

func (this *exchangeVerifyModule) Run(closeSig chan bool) {
	//panic("implement me")
}

func (this *exchangeVerifyModule) OnDestroy() {
	this.BaseModule.Destroy()
}
