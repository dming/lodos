package basemodule

import (
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/rpc/pb"
)

type moduleSession struct {
	id        string
	app       module.AppInterface
	mType     string
	rpcClient rpc.RPCClient
}

func NewModuleSession(app module.AppInterface, id string, Type string, client rpc.RPCClient) module.ModuleSession {
	s := &moduleSession{
		app: app,
		id: id,
		mType: Type,
		rpcClient: client,
	}
	return s
}

func (s *moduleSession) GetApp () module.AppInterface {
	return s.app
}

func (s *moduleSession) GetId () string {
	return s.id
}

func (s *moduleSession) GetType () string {
	return s.mType
}

func (s *moduleSession) GetClient () rpc.RPCClient {
	return s.rpcClient
}

func (s *moduleSession) Call(_func string, params ...interface{}) ([]interface{}, error) {
	return s.rpcClient.Call(_func, params...)
}
func (s *moduleSession) SyncCall(_func string, params ...interface{}) (chan rpcpb.ResultInfo, error) {
	return s.rpcClient.SyncCall(_func, params...)
}
func (s *moduleSession) CallNR(_func string, params ...interface{}) (err error) {
	return s.rpcClient.CallNR(_func, params...)
}
func (s *moduleSession) CallArgs(_func string, ArgsType []string, Args [][]byte) ([]interface{}, error) {
	return s.rpcClient.CallArgs(_func, ArgsType, Args)
}
func (s *moduleSession) SyncCallArgs(_func string, ArgsType []string, Args [][]byte) (chan rpcpb.ResultInfo, error) {
	return s.rpcClient.SyncCallArgs(_func, ArgsType, Args)
}
func (s *moduleSession) CallArgsNR(_func string, ArgsType []string, Args [][]byte) (err error) {
	return s.rpcClient.CallArgsNR(_func, ArgsType, Args)
}