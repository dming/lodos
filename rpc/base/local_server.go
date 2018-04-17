package baserpc

import (
	"github.com/dming/lodos/rpc"
	"sync"
	"fmt"
	"github.com/dming/lodos/rpc/pb"
	"github.com/dming/lodos/log"
)

type localServer struct {
	call_chan chan rpc.CallInfo // chan to rpc server
	local_chan chan rpc.CallInfo // chan to local client
	done chan error
	isClose bool
	lock *sync.Mutex
} 

func NewLocalServer(call_chan chan rpc.CallInfo) (*localServer, error) {
	server := new(localServer)
	server.call_chan = call_chan
	server.local_chan = make(chan rpc.CallInfo, 1)
	server.isClose = false
	server.lock = new(sync.Mutex)
	go server.on_request_handle(server.local_chan)
	return server, nil
}

func (s *localServer) IsClose() bool {
	return s.isClose;
}

func (s *localServer) WriteToRpcServer(callInfo rpc.CallInfo) (success bool) {
	defer func() {
		if recover() != nil {
			log.Error("s.call_chan is closed")
			success = false
		}
	}()

	if s.isClose {
		return false
	}
	s.call_chan <- callInfo
	return true
}

//stop
func (s *localServer) StopConsume() error {
	s.lock.Lock()
	s.isClose = true
	s.lock.Unlock()
	return nil
}

/**
注销消息队列
*/
func (s *localServer) Shutdown() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()
	return nil
}

func (s *localServer) Callback (callInfo rpc.CallInfo) error {
	replyTo := callInfo.Props["reply_to"].(chan rpcpb.ResultInfo)
	replyTo <- callInfo.Result
	return nil
}

func (s *localServer) on_request_handle(local_chan <-chan rpc.CallInfo) {
	for {
		select {
		case callInfo, ok := <-local_chan:
			if !ok {
				local_chan = nil
			} else {
				callInfo.Agent = s
				if s.WriteToRpcServer(callInfo) {
					log.Warning("rpc request [%s] fail : [%s]", callInfo.RpcInfo.Fn, callInfo.RpcInfo.Cid)
				}
			}
		}
		if local_chan == nil {
			break
		}
	}
}

func (s *localServer) SafeCallback(local_chan chan rpcpb.ResultInfo, resultInfo rpcpb.ResultInfo) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	// assume ch != nil here.
	local_chan <- resultInfo
	return false
}
func (s *localServer) SafeSend(call_chan chan rpc.CallInfo, callInfo rpc.CallInfo) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	// assume ch != nil here.
	call_chan <- callInfo
	return false
}

