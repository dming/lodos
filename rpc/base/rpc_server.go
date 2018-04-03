package baserpc

import (
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/rpc"
	"sync"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/log"
	"fmt"
	"time"
	"reflect"
	"github.com/opentracing/opentracing-go"
	"github.com/dming/lodos/gate"
	"github.com/dming/lodos/rpc/utils"
	"github.com/dming/lodos/rpc/pb"
	"runtime"
)

type RPCServer struct {
	module 				module.Module
	app					module.AppInterface
	functions 			map[string]rpc.FuncInfo

	localServer 		*localServer
	remoteServer 		*amqpServer
	redisServer 		*redisServer

	mq_chan 			chan rpc.CallInfo //接收到请求信息的队列
	callback_chan 		chan rpc.CallInfo //信息处理完成的队列
	mq_chan_done 		chan error
	callback_chan_done 	chan error
	wg 					sync.WaitGroup  //任务阻塞

	listener 			rpc.RPCListener
	control 			rpc.GoroutineControl 	//控制模块可同时开启的最大协程数

	executing 			int64					//正在执行的goroutine数量
	ch 					chan int				//控制模块可同时开启的最大协程数
}

func NewRPCServer(app module.AppInterface, module module.Module) (rpc.RPCServer, error) {
	server := new(RPCServer)
	server.app = app
	server.module = module
	server.functions = make(map[string]rpc.FuncInfo)
	server.mq_chan = make(chan rpc.CallInfo)
	server.callback_chan = make(chan rpc.CallInfo, 1)
	server.mq_chan_done = make(chan error)
	server.callback_chan_done = make(chan error)
	//server.ch = make(chan int, app.GetSettings().Rpc.MaxCoroutine) //need to be complete
	server.SetGoroutineControl(server)

	//create local rpc service
	local_server, err := NewLocalServer(server.mq_chan)
	if err != nil {
		log.Error("LocalServer Dial : %s", err)
	}
	server.localServer = local_server

	go server.on_call_handle(server.mq_chan, server.callback_chan, server.mq_chan_done)
	go server.on_callback_handle(server.callback_chan, server.callback_chan_done)

	return server, nil
}

func (this *RPCServer) Wait() error {
	// 如果ch满了则会处于阻塞，从而达到限制最大协程的功能
	this.ch <- 1
	return nil
}
func (this *RPCServer) Finish() {
	// 完成则从ch推出数据
	<-this.ch
}

//create remote rpc service
func (s *RPCServer) NewRabbitmqRpcServer (info *conf.Rabbitmq) (err error) {
	remoteServer, err := NewAMQPServer(info, s.mq_chan)
	if err != nil {
		log.Error("AMQP Server Dial : %s", err)
	}
	s.remoteServer = remoteServer
	return
}
//create redis rpc service
func (s *RPCServer) NewRedisRpcServer(info *conf.Redis) (err error) {
	redisServer, err := NewRedisServer(info, s.mq_chan)
	if err != nil {
		log.Error("Redis server Dial : %s", err)
	}
	s.redisServer = redisServer
	return
}

func (s *RPCServer) SetListener(listener rpc.RPCListener) {
	s.listener = listener
}

func (s *RPCServer) SetGoroutineControl(control rpc.GoroutineControl) {
	s.control = control
}

func (s *RPCServer) GetLocalRpcServer() rpc.LocalServer {
	return s.localServer
}

func (s *RPCServer) GetExecuting() int64 {
	return s.executing
}

//
func (s *RPCServer) Register(id string, fn interface{}) {
	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id));
	}

	s.functions[id] = *&rpc.FuncInfo{
		Function: fn,
		Goroutine: false,
	}
}

//
func (s *RPCServer) RegisterGo(id string, fn interface{}) {
	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id));
	}

	s.functions[id] = *&rpc.FuncInfo{
		Function: fn,
		Goroutine: true,
	}
}

//need check
func (s *RPCServer) Done() (err error) {
	//stop
	if s.remoteServer != nil {
		err = s.remoteServer.StopConsume()
	}
	if s.localServer != nil {
		err = s.localServer.StopConsume()
	}
	if s.redisServer != nil {
		err = s.redisServer.StopConsume()
	}

	s.wg.Wait()
	s.mq_chan_done <- nil
	s.callback_chan_done <- nil
	close(s.callback_chan) // 关闭结果发送队列

	if s.remoteServer != nil {
		err = s.remoteServer.Shutdown()
	}
	if s.localServer != nil {
		err = s.localServer.Shutdown()
	}
	if s.redisServer != nil {
		err = s.redisServer.Shutdown()
	}

	return
}


//处理结果，使用回调
func (s *RPCServer) on_callback_handle(callback_chan <-chan rpc.CallInfo, done chan error) {
	for {
		select {
		case callInfo, ok := <-callback_chan:
			if !ok {
				callback_chan = nil
			} else {
				if callInfo.RpcInfo.Reply {
					// need to be reply
					//需要回复的才回复
					err := callInfo.Agent.(rpc.MQServer).Callback(callInfo)
					if err != nil {
						log.Warning("rpc callback err : %\n%s", err.Error())
					}
				} else {
				}
			}
		case <-done :
			goto EForEnd
		}
		if callback_chan == nil {
			goto EForEnd
		}
	}
EForEnd:
}


/**
接收请求信息
*/
func (s *RPCServer) on_call_handle(call_chan <-chan rpc.CallInfo, callback_chan chan rpc.CallInfo, done chan error) {
	for {
		select {
		case callInfo, ok := <-call_chan:
			if !ok {
				goto ForEnd
			} else {
				if callInfo.RpcInfo.Expired < (time.Now().UnixNano() / 1000000) {
					//请求超时了,无需再处理
					if s.listener != nil {
						//s.listener.OnTimeOut(callInfo.RpcInfo.Fn, callInfo.RpcInfo.Expired)
					} else {
						log.Warning("timeout: This is Call", s.module.GetType(), callInfo.RpcInfo.Fn, callInfo.RpcInfo.Expired, time.Now().UnixNano()/1000000)
					}
				} else {
					s.runFunc(callInfo, callback_chan)
				}
			}
		case <-done:
			goto ForEnd
		}
	}
ForEnd:
}

func (s *RPCServer) runFunc(callInfo rpc.CallInfo, callback_chan chan<- rpc.CallInfo) {

	_errorCallback := func(Cid string, Error string, span opentracing.Span, traceid string) {
		//异常日志都应该打印
		log.Error("[%s] %s rpc func(%s) error:\n%s", traceid, s.module.GetType(), callInfo.RpcInfo.Fn, Error)
		resultInfo := rpcpb.NewResultInfo(Cid, Error, nil, nil)
		callInfo.Result = *resultInfo
		callback_chan <- callInfo
		if span != nil {
			span.LogEventWithPayload("Error", Error)
		}
		if s.listener != nil {
			//s.listener.OnError(callInfo.RpcInfo.Fn, &callInfo, fmt.Errorf(Error))
		}
	}
	// defer
	defer func() {
		if r := recover(); r != nil {
			var rn = ""
			switch r.(type) {

			case string:
				rn = r.(string)
			case error:
				rn = r.(error).Error()
			}
			log.Error("recover", rn)
			_errorCallback(callInfo.RpcInfo.Cid, rn, nil, "")
		}
	}()

	functionInfo, ok := s.functions[callInfo.RpcInfo.Fn]
	if !ok {
		_errorCallback(callInfo.RpcInfo.Cid, fmt.Sprintf("Remote function(%s) not found", callInfo.RpcInfo.Fn), nil, "")
		return
	}
	_func := functionInfo.Function
	args := callInfo.RpcInfo.Args
	argsType := callInfo.RpcInfo.ArgsType
	f := reflect.ValueOf(_func)
	if len(args) != f.Type().NumIn() {
		_errorCallback(callInfo.RpcInfo.Cid, fmt.Sprintf("The number of args %s is not adapted.%s", args, f.String()), nil, "")
		return
	}

	_runFunc := func() {
		s.wg.Add(1)
		s.executing++
		var span opentracing.Span = nil
		var tradeId string = ""
		defer func() {
			if r := recover(); r != nil {
				var rn = ""
				switch r.(type) {

				case string:
					rn = r.(string)
				case error:
					rn = r.(error).Error()
				}
				buf := make([]byte, 1024)
				l := runtime.Stack(buf, false)
				errstr := string(buf[:l])
				allError := fmt.Sprintf("%s rpc func(%s) error %s\n ----Stack----\n%s", s.module.GetType(), callInfo.RpcInfo.Fn, rn, errstr)
				log.Error(allError)
				_errorCallback(callInfo.RpcInfo.Cid, allError, span, tradeId)
			}

			if span != nil {
				span.Finish()
			}
			s.wg.Add(-1)
			s.executing--
			if s.control != nil {
				s.control.Finish()
			}
		}()

		//exec_time := time.Now().UnixNano()

		//var session gate.Session = nil
		var in []reflect.Value

		if len(argsType) > 0 {
			in = make([]reflect.Value, len(args))
			for k, v := range argsType {
				v, err := argsutil.Bytes2Args(s.app, v, args[k])
				if err != nil {
					_errorCallback(callInfo.RpcInfo.Cid, err.Error(), span, tradeId)
					return
				}
				//switch v2 := v.(type) {
				switch v.(type) {
				case gate.Session:
					//try to load Span
					//span = v2.LoadSpan(fmt.Sprintf("%s/%s", s.module.GetType(), callInfo.RpcInfo.Fn))
					/*if span != nil {
						span.SetTag("UserId", v2.GetUserid())
						span.SetTag("Func", callInfo.RpcInfo.Fn)
					}
					session = v2
					traceid = session.TracId()*/
					in[k] = reflect.ValueOf(v)
				case nil:
					in[k] = reflect.Zero(f.Type().In(k))//设定为f的第k个参数的零值
				default:
					in[k] = reflect.ValueOf(v)
				}
			}
		} //init in here

		if s.listener != nil {
			/*
			errs := s.listener.BeforeHandle(callInfo.RpcInfo.Fn, session, &callInfo)
			if errs != nil {
				_errorCallback(callInfo.RpcInfo.Cid, errs.Error(), span, traceid)
				return
			}*/
		}

		out := f.Call(in)
		var rs []interface{}

		if len(out) < 1 {
			_errorCallback(callInfo.RpcInfo.Cid, "the number of result output is less than 1.", span, tradeId)
			return
		}
		if len(out) > 0 {
			rs = make([]interface{}, len(out))
			for i, v := range out {
				rs[i] = v.Interface()
			}
		}
		var argsType = make([]string, len(out)-1)
		var args =  make([][]byte, len(out)-1)
		//argsType, args, err := argsutil.ArgsTypeAnd2Bytes(s.app, rs[0])
		for i := 0; i < len(out) - 1; i++ {
			var err error = nil
			argsType[i], args[i], err = argsutil.ArgsTypeAnd2Bytes(s.app, rs[i])
			if err != nil {
				_errorCallback(callInfo.RpcInfo.Cid, err.Error(), span, tradeId)
				return
			}
		}
		/*
		// 错误信息应该是string或者error格式的
		var errStr string = ""
		switch rs[len(out) - 1].(type) {
		case types.Nil:
			errStr = ""
		case string :
			errStr = rs[len(out) - 1].(string);
		case error :
			errStr = rs[len(out) - 1].(error).Error()
		default:
			_errorCallback(callInfo.RpcInfo.Cid, fmt.Sprintf("the error.(type) is not as string neither error"), span, tradeId)
			return
		}*/
		resultInfo := rpcpb.NewResultInfo(
			callInfo.RpcInfo.Cid,
			rs[len(out) - 1].(string),
			argsType,
			args,
		)
		callInfo.Result = *resultInfo
		callback_chan <- callInfo
		/*if span != nil {
			span.LogEventWithPayload("Result.Type", argsType)
			span.LogEventWithPayload("Result", string(args))
		}
		if s.app.GetSettings().Rpc.LogSuccess {
			log.Info("[%s] %s rpc func(%s) exec_time(%s) success", traceid, s.module.GetType(), callInfo.RpcInfo.Fn, s.timeConversion(time.Now().UnixNano()-exec_time))
		}
		if s.listener != nil {
			s.listener.OnComplete(callInfo.RpcInfo.Fn, &callInfo, resultInfo, time.Now().UnixNano()-exec_time)
		}*/
	}

	if s.control != nil {
		//协程数量达到最大限制
		s.control.Wait()
	}

	if functionInfo.Goroutine {
		go _runFunc()
	} else {
		_runFunc()
	}
}

//时间转换
func (s *RPCServer) timeConversion(ns int64) string {
	if (ns / 1000) < 1 {
		return fmt.Sprintf("%d ns", (ns))
	} else if 1 < (ns/int64(1000)) && (ns/int64(1000)) < 1000 {
		return fmt.Sprintf("%.2f μs", float32(ns/int64(1000)))
	} else if 1 < (ns/int64(1000*1000)) && (ns/int64(1000*1000)) < 1000 {
		return fmt.Sprintf("%.2f ms", float32(ns/int64(1000*1000)))
	} else if 1 < (ns/int64(1000*1000*1000)) && (ns/int64(1000*1000*1000)) < 1000 {
		return fmt.Sprintf("%.2f s", float32(ns/int64(1000*1000*1000)))
	} else if 1 < (ns/int64(1000*1000*1000*60)) && (ns/int64(1000*1000*1000*60)) < 1000 {
		return fmt.Sprintf("%.2f m", float32(ns/int64(1000*1000*1000*60)))
	} else if 1 < (ns/int64(1000*1000*1000*60*60)) && (ns/int64(1000*1000*1000*60*60)) < 1000 {
		return fmt.Sprintf("%.2f m", float32(ns/int64(1000*1000*1000*60*60)))
	} else {
		return fmt.Sprintf("%d ns", (ns))
	}
}
















/*
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
		log.Error("can not find the function name : %v\n", ci.Id)
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
*/

