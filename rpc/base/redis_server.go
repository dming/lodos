package baserpc

import (
	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/log"
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/rpc/pb"
	"github.com/dming/lodos/utils"
	"runtime"
)

type redisServer struct {
	call_chan chan rpc.CallInfo // chan to rpc server
	url       string
	info      *conf.Redis
	queueName string
	done      chan error
	pool      redis.Conn
	closed    bool
}

func NewRedisServer(info *conf.Redis, call_chan chan rpc.CallInfo) (*redisServer, error) {
	var queueName = info.Queue
	var url = info.Uri
	server := new(redisServer)
	server.call_chan = call_chan
	server.url = url
	server.info = info
	server.queueName = queueName
	server.done = make(chan error)
	server.closed = false
	go server.on_request_handle(server.done)

	return server, nil
	//log.Printf("shutting down")
	//
	//if err := c.Shutdown(); err != nil {
	//	log.Fatalf("error during shutdown: %s", err)
	//}
}

/**
停止接收请求
*/
func (s *redisServer) StopConsume() error {
	s.closed = true
	if s.pool != nil {
		return s.pool.Close()
	}
	return nil
}

/**
注销消息队列
*/
func (s *redisServer) Shutdown() error {
	s.closed = true
	if s.pool != nil {
		return s.pool.Close()
	}
	return nil
}

func (s *redisServer) Callback(callinfo rpc.CallInfo) error {
	body, _ := s.MarshalResult(callinfo.Result)
	return s.response(callinfo.Props, body)
}

/**
消息应答
*/
func (s *redisServer) response(props map[string]interface{}, body []byte) error {
	pool := utils.GetRedisFactory().GetPool(s.info.Uri).Get()
	defer pool.Close()
	var err error
	_, err = pool.Do("lpush", props["reply_to"].(string), body)
	if err != nil {
		log.Warning("Publish: %s", err)
		return err
	}
	return nil
}

/**
接收请求信息
*/
func (s *redisServer) on_request_handle(done chan error) {
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
			log.Error("%s\n ----Stack----\n%s", rn, errstr)
		}
	}()
	for !s.closed {
		s.pool = utils.GetRedisFactory().GetPool(s.info.Uri).Get()
		result, err := s.pool.Do("brpop", s.queueName, 0)
		s.pool.Close()
		if err == nil && result != nil {
			rpcInfo, err := s.Unmarshal(result.([]interface{})[1].([]byte))
			if err == nil {
				callInfo := &rpc.CallInfo{
					RpcInfo: *rpcInfo,
				}
				callInfo.Props = map[string]interface{}{
					"reply_to": callInfo.RpcInfo.ReplyTo,
				}

				callInfo.Agent = s //设置代理为AMQPServer

				s.call_chan <- *callInfo
			} else {
				log.Error("error ", err)
			}
		} else if err != nil {
			log.Error("error %s", err.Error())
		}
	}
	log.Debug("finish on_request_handle")
}

func (s *redisServer) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var rpcInfo rpcpb.RPCInfo
	err := proto.Unmarshal(data, &rpcInfo)
	if err != nil {
		return nil, err
	} else {
		return &rpcInfo, err
	}

	panic("bug")
}

// goroutine safe
func (s *redisServer) MarshalResult(resultInfo rpcpb.ResultInfo) ([]byte, error) {
	//log.Error("",map2)
	b, err := proto.Marshal(&resultInfo)
	return b, err
}
