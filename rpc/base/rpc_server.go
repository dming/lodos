package baserpc

import (
	"reflect"
	"fmt"
	"errors"
	"sync"
	"rpc"
)

type ServerInterface interface {
	GetChanCall()
	Register()
	RegisterGo()
	Run()
	Exec()
	exec()
	ret()
	Close()
}

func NewServer() *server {
	s := new(server)
	s.functions = make(map[string]*rpc.FuncInfo)
	s.ChanCall = make(chan *rpc.CallInfo, 10000)
	s.ChanDone = make(chan error, 1)
	s.executing = 0
	go s.Run(s.ChanCall, s.ChanDone)
	return s
}

type server struct {
	functions map[string]*rpc.FuncInfo
	ChanCall chan *rpc.CallInfo //CallInfo 包括 name, chanback, error, 其中chanback is back to localserver or mqserver
	ChanDone chan error
	MqServer rpc.MqServer

	executing int64
	wg sync.WaitGroup

}

func (s *server) GetChanCall() (chan *rpc.CallInfo) {
	return s.ChanCall
}

func (s *server) AttachMqServer(mqServer rpc.MqServer)  {
	s.MqServer = mqServer
}


// the function should be as bellow :
// func funcName (inputs) (outputs, error)
// PS: error in output is require
func (s *server) Register(id string, f interface{}) {
	// f should be valid
	if reflect.TypeOf(f).Kind() != reflect.Func {
		panic(fmt.Sprintln("the arg f is not the type of Function."))
	}

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	fi := new(rpc.FuncInfo)
	fi.Fn = reflect.ValueOf(f)
	fi.IsGo = false
	s.functions[id] = fi
}

func (s *server) RegisterGo(id string, f interface{}) {
	// f should be valid
	if reflect.TypeOf(f).Kind() != reflect.Func {
		panic(fmt.Sprintln("the arg f is not the type of Function."))
	}

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	fi := new(rpc.FuncInfo)
	fi.Fn = reflect.ValueOf(f)
	fi.IsGo = true
	s.functions[id] = fi
}


func (s *server) Run(chanCall chan *rpc.CallInfo, chanDone chan error)  {
	for {
		select {
		case ci, ok := <- chanCall:
			if !ok {
				chanCall = nil
			} else {
				if false {
					// something not good
				} else {
					s.Exec(ci)
				}
			}
		}
		if chanCall == nil {
			chanDone <- nil
			break
		}
	}
}



func (s *server) Exec(ci *rpc.CallInfo) {
	// check callinfo.fuctionName
	fi, ok := s.functions[ci.Id]
	if !ok {
		fmt.Printf("cant find the function name : %v\n", ci.Id)
		s := "can not find the function name: " + ci.Id
		err := errors.New(s)
		ri := new(rpc.RetInfo)
		ri.Err = err
		ci.ChanRet <- ri

		return
	}

	defer func() {
		if r := recover(); r != nil {
			//s.ret(ci, &ChanRetInfo{err: fmt.Errorf("%v", r)})
		}
	}()

	// if sync, exec(f reflect.value)
	if !fi.IsGo {
		s.exec(fi.Fn, ci)
	} else {
		go s.exec(fi.Fn, ci)
	}

}

func (s *server) exec(fn reflect.Value, ci *rpc.CallInfo) (err error) {
	//f, ok := s.functions[ci.id]

	in := make([]reflect.Value, len(ci.Args))
	for k, arg := range ci.Args {
		in[k] = reflect.ValueOf(arg)
	}

	retValues := fn.Call(in)
	ret := make([]interface{}, len(retValues))
	for i, v := range retValues {
		ret[i] = v.Interface()
	}

	// fn must be execAble..guarantee by client
	return s.ret(ci, &rpc.RetInfo{Ret: ret})
}

func (s *server) ret(ci *rpc.CallInfo, ri *rpc.RetInfo) (err error) {

	if ci.ChanRet == nil && ci.ReplyTo == "" {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	// ri only has Ret right now
	if len(ri.Ret) > 1 &&  ri.Ret[1] != nil {
		ri.Err = ri.Ret[1].(error)
	}
	temp := make([]interface{}, 1)
	temp[0] = ri.Ret[0]
	ri.Ret = temp

	if (ci.ReplyTo == "") {
		ci.ChanRet <- ri
	} else {
		//return to mqserver
		//s.MqServer.ChanRet <- ri
		ri.ReplyTo = ci.ReplyTo
		ri.Flag = ci.Flag
		ci.ChanRet <- ri

	}

	return
}

func (s *server) Close() error {
	close(s.ChanCall)

	var err error
	for ci := range s.ChanCall {
		errC := s.ret(ci, &rpc.RetInfo{
			Err: errors.New("chanrpc server closed"),
		})
		if errC != nil {
			err = errC
		}
	}

	if s.MqServer != nil {
		errM := s.MqServer.Shutdown()
		if errM != nil {
			err = errM
		}
	}

	return err
}
