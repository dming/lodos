package auth

import "github.com/casbin/casbin"

type auth struct {
	enforcer *casbin.Enforcer
}

func (a *auth)OnInit() {
	a.enforcer = casbin.NewEnforcer("authk", "authv")
}
