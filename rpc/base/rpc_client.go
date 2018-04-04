package baserpc

import (
	"github.com/dming/lodos/rpc/pb"
	"github.com/dming/lodos/rpc/argsutil"
	"fmt"
	"github.com/dming/lodos/gate"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/utils/uuid"
	"github.com/golang/protobuf/proto"
	"time"
	"github.com/dming/lodos/log"
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/rpc"
)

type ClientCallInfo struct {
	correlation_id string
	timeout        int64 //超时
	call_chan           chan rpcpb.ResultInfo
}


type rpcClient struct {
	app           module.AppInterface
	serverId      string
	remote_client *amqpClient
	local_client  *localClient
	redis_client  *redisClient
}

func NewRPCClient(app module.AppInterface, serverId string) (rpc.RPCClient, error) {
	client := new(rpcClient)
	client.serverId = serverId
	client.app = app
	return client, nil
}

//only init once
func (c *rpcClient) NewRabbitmqRpcClient(info *conf.Rabbitmq) (err error) {
	//创建本地连接
	if info != nil && c.remote_client == nil {
		c.remote_client, err = NewAMQPClient(info)
		if err != nil {
			log.Error("RabbitmqRpcClient Dial : %s", err)
		}
	}
	return
}

//only init once
func (c *rpcClient) NewLocalRpcClient(server rpc.RPCServer) (err error) {
	//创建本地连接
	if server != nil && server.GetLocalRpcServer() != nil && c.local_client == nil {
		c.local_client, err = NewLocalClient(server.GetLocalRpcServer())
		if err != nil {
			log.Error("LocalRpcClient Dial: %s", err)
		}
	}
	return
}

//only init once
func (c *rpcClient) NewRedisRpcClient(info *conf.Redis) (err error) {
	//创建本地连接
	if info != nil && c.redis_client == nil {
		c.redis_client, err = NewRedisClient(info)
		if err != nil {
			log.Error("RedisRpcClient Dial: %s", err)
		}
	}
	return
}

func (c *rpcClient) Done() (err error) {
	if c.remote_client != nil {
		err = c.remote_client.Done()
	}
	if c.redis_client != nil {
		err = c.redis_client.Done()
	}
	if c.local_client != nil {
		err = c.local_client.Done()
	}
	return
}

func (c *rpcClient) CallArgs(_func string, ArgsType []string, args [][]byte) ([]interface{}, error) {
	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &rpcpb.RPCInfo{
		Fn:       *proto.String(_func),
		Reply:    *proto.Bool(true),
		Expired:  *proto.Int64((time.Now().UTC().Add(time.Second * time.Duration(c.app.GetSettings().RPC.RpcExpired)).UnixNano()) / 1000000),
		Cid:      *proto.String(correlation_id),
		Args:     args,
		ArgsType: ArgsType,
	}
	if c.app.GetSettings().RPC.Log {
		log.Info("request %s rpc func(%s) [%s]", c.serverId, _func, correlation_id)
	}

	callInfo := &rpc.CallInfo{
		RpcInfo: *rpcInfo,
	}
	callback_chan := make(chan rpcpb.ResultInfo, 1)
	var err error
	//优先使用本地rpc
	if c.local_client != nil {
		err = c.local_client.Call(*callInfo, callback_chan)
	} else if c.remote_client != nil {
		err = c.remote_client.Call(*callInfo, callback_chan)
	} else if c.redis_client != nil {
		err = c.redis_client.Call(*callInfo, callback_chan) // 事实上，redisClient 和其他的RPC应该用途不一样的吧？？
	} else {
		return nil, fmt.Errorf("rpc service (%s) connection failed", c.serverId)
	}
	if err != nil {
		return nil, err
	}
	resultInfo, ok := <-callback_chan
	if !ok {
		return nil, fmt.Errorf("rpc client closed")
	}
	if len(resultInfo.Results) != len(resultInfo.ResultsType) {
		resultInfo.Error += " len(resultInfo.Results) != len(resultInfo.ResultsType) "
		return nil, fmt.Errorf(resultInfo.Error)
	}
	results := make([]interface{}, len(resultInfo.Results))
	for i := 0; i < len(resultInfo.Results); i++ {
		results[i], err = argsutil.Bytes2Args(c.app, resultInfo.ResultsType[i], resultInfo.Results[i])
		if err != nil {
			return nil, fmt.Errorf(resultInfo.Error + err.Error())
		}
	}
	if c.app.GetSettings().RPC.Log {
		log.Info("response %s rpc func(%s) [%s]", c.serverId, _func, correlation_id)
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (c *rpcClient) SyncCallArgs(_func string, ArgsType []string, args [][]byte) (chan rpcpb.ResultInfo, error) {
	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &rpcpb.RPCInfo{
		Fn:       *proto.String(_func),
		Reply:    *proto.Bool(true),
		Expired:  *proto.Int64((time.Now().UTC().Add(time.Second * time.Duration(c.app.GetSettings().RPC.RpcExpired)).UnixNano()) / 1000000),
		Cid:      *proto.String(correlation_id),
		Args:     args,
		ArgsType: ArgsType,
	}
	if c.app.GetSettings().RPC.Log {
		log.Info("request %s rpc func(%s) [%s]", c.serverId, _func, correlation_id)
	}

	callInfo := &rpc.CallInfo{
		RpcInfo: *rpcInfo,
	}
	callback_chan := make(chan rpcpb.ResultInfo, 1)
	var err error
	//优先使用本地rpc
	if c.local_client != nil {
		err = c.local_client.Call(*callInfo, callback_chan)
	} else if c.remote_client != nil {
		err = c.remote_client.Call(*callInfo, callback_chan)
	} else if c.redis_client != nil {
		err = c.redis_client.Call(*callInfo, callback_chan) // 事实上，redisClient 和其他的RPC应该用途不一样的吧？？
	} else {
		return nil, fmt.Errorf("rpc service (%s) connection failed", c.serverId)
	}
	if err != nil {
		return nil, err
	} else {
		return callback_chan, nil
	}

}

func (c *rpcClient) CallArgsNR(_func string, ArgsType []string, args [][]byte) (err error) {
	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &rpcpb.RPCInfo{
		Fn:       *proto.String(_func),
		Reply:    *proto.Bool(false),
		Expired:  *proto.Int64((time.Now().UTC().Add(time.Second * time.Duration(c.app.GetSettings().RPC.RpcExpired)).UnixNano()) / 1000000),
		Cid:      *proto.String(correlation_id),
		Args:     args,
		ArgsType: ArgsType,
	}
	if c.app.GetSettings().RPC.Log {
		log.Info("request %s rpc(nr) func(%s) [%s]", c.serverId, _func, correlation_id)
	}
	callInfo := &rpc.CallInfo{
		RpcInfo: *rpcInfo,
	}
	//优先使用本地rpc
	if c.local_client != nil {
		err = c.local_client.CallNR(*callInfo)
	} else if c.remote_client != nil {
		err = c.remote_client.CallNR(*callInfo)
	} else if c.redis_client != nil {
		err = c.redis_client.CallNR(*callInfo)
	} else {
		return fmt.Errorf("rpc service (%s) connection failed", c.serverId)
	}
	return nil
}

/**
消息请求 需要回复
*/
func (c *rpcClient) Call(_func string, params ...interface{}) ([]interface{}, error) {
	var argsType []string = make([]string, len(params))
	var args [][]byte = make([][]byte, len(params))
	for k, param := range params {
		var err error = nil
		argsType[k], args[k], err = argsutil.ArgsTypeAnd2Bytes(c.app, param)
		if err != nil {
			return nil, fmt.Errorf("args[%d] error %s", k, err.Error())
		}

		switch v2 := param.(type) { //多选语句switch
		case gate.Session:
			//如果参数是这个需要拷贝一份新的再传 ???? 用途是什么？？
			param = v2.(gate.Session).Clone()
		}
	}
	return c.CallArgs(_func, argsType, args)
}

func (c *rpcClient) SyncCall(_func string, params ...interface{}) (chan rpcpb.ResultInfo, error) {
	var argsType []string = make([]string, len(params))
	var args [][]byte = make([][]byte, len(params))
	for k, param := range params {
		var err error = nil
		argsType[k], args[k], err = argsutil.ArgsTypeAnd2Bytes(c.app, param)
		if err != nil {
			return nil, fmt.Errorf("args[%d] error %s", k, err.Error())
		}

		switch v2 := param.(type) { //多选语句switch
		case gate.Session:
			//如果参数是这个需要拷贝一份新的再传 ???? 用途是什么？？
			param = v2.(gate.Session).Clone()
		}
	}
	return c.SyncCallArgs(_func, argsType, args)
}


/**
消息请求 不需要回复
*/
func (c *rpcClient) CallNR(_func string, params ...interface{}) (err error) {
	var ArgsType []string = make([]string, len(params))
	var args [][]byte = make([][]byte, len(params))
	for k, param := range params {
		ArgsType[k], args[k], err = argsutil.ArgsTypeAnd2Bytes(c.app, param)
		if err != nil {
			return fmt.Errorf("args[%d] error %s", k, err.Error())
		}

		switch v2 := param.(type) { //多选语句switch
		case gate.Session:
			//如果参数是这个需要拷贝一份新的再传
			param = v2.Clone()
		}
	}
	return c.CallArgsNR(_func, ArgsType, args)
}

func (c *rpcClient) ResolveCallbackChan(callback_chan chan rpcpb.ResultInfo, _func string, cid string) ([]interface{}, error) {
	var err error
	resultInfo, ok := <-callback_chan
	if !ok {
		return nil, fmt.Errorf("client closed")
	}
	if len(resultInfo.Results) != len(resultInfo.ResultsType) {
		resultInfo.Error += " len(resultInfo.Results) != len(resultInfo.ResultsType) "
		return nil, fmt.Errorf(resultInfo.Error)
	}
	results := make([]interface{}, len(resultInfo.Results))
	for i := 0; i < len(resultInfo.Results); i++ {
		results[i], err = argsutil.Bytes2Args(c.app, resultInfo.ResultsType[i], resultInfo.Results[i])
		if err != nil {
			return nil, fmt.Errorf(resultInfo.Error + err.Error())
		}
	}
	if c.app.GetSettings().RPC.Log {
		log.Info("response %s rpc func(%s) [%s]", c.serverId, _func, cid)
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}
