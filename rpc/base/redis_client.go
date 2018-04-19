package baserpc

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/log"
	"github.com/dming/lodos/rpc/pb"
	"github.com/dming/lodos/utils"
	"time"
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/db/base"
)

type redisClient struct {
	//callInfos map[string]*ClientCallInfo
	callInfos         utils.ConcurrentMap
	redisInfo         *conf.Redis
	callQueueName     string
	callbackQueueName string
	done              chan error
	timeout_done      chan error
	pool              redis.Conn
	isClose           bool
	ch                chan int //控制一个RPC 客户端可以同时等待的最大消息量，
	// 如果等待的请求过多代表rpc server请求压力大，
	// 就不要在往消息队列发消息了,客户端先阻塞
	//cmutex            sync.Mutex //操作callinfos的锁
}

func createQueueName() string {
	//return "callbackQueueName"
	return fmt.Sprintf("callbackQueueName:%d", time.Now().Nanosecond())
}
func NewRedisClient(info *conf.Redis) (client *redisClient , err error) {
	client = new(redisClient)
	client.callInfos = utils.New()
	client.redisInfo = info
	client.callbackQueueName = createQueueName()
	client.callQueueName = info.RPCQueue
	client.done = make(chan error)
	client.timeout_done = make(chan error)
	client.isClose = false
	client.ch = make(chan int, 100)
	pool := basedb.GetRedisFactory().GetPool(info.RPCUri).Get()
	defer pool.Close()
	//_, errs:=pool.Do("EXPIRE", client.callbackQueueName, 60)
	//if errs != nil {
	//	log.Warning(errs.Error())
	//}
	go client.on_response_handle(client.done)
	go client.on_timeout_handle(client.timeout_done) //处理超时请求的协程
	return client, nil
}


func (c *redisClient ) Wait() error {
	// 如果ch满了则会处于阻塞，从而达到限制最大协程的功能
	//c.ch <- 1
	return nil
}
func (c *redisClient ) Finish() {
	// 完成则从ch推出数据
	//<-c.ch
}

func (c *redisClient ) Done() (err error) {
	pool := basedb.GetRedisFactory().GetPool(c.redisInfo.RPCUri).Get()
	defer pool.Close()
	//删除临时通道
	pool.Do("DEL", c.callbackQueueName)
	c.isClose = true
	c.timeout_done <- nil
	close(c.timeout_done)
	//err = c.psc.Close()
	//清理 callInfos 列表
	if c.pool != nil {
		c.pool.Close()
	}
	for tuple := range c.callInfos.IterBuffered() {
		//关闭管道
		close(tuple.Val.(*ClientCallInfo).call_chan)
		//从Map中删除
		c.callInfos.Remove(tuple.Key)
	}
	return
}

/**
消息请求
*/
func (c *redisClient ) Call(callInfo rpc.CallInfo, callback_chan chan rpcpb.ResultInfo) error {
	pool := basedb.GetRedisFactory().GetPool(c.redisInfo.RPCUri).Get()
	defer pool.Close()
	var err error
	if c.isClose {
		return fmt.Errorf("RedisClient is isClose")
	}
	callInfo.RpcInfo.ReplyTo = c.callbackQueueName
	correlation_id := callInfo.RpcInfo.Cid

	ClientCallInfo := &ClientCallInfo{
		correlation_id: correlation_id,
		call_chan:           callback_chan,
		timeout:        callInfo.RpcInfo.Expired,
	}

	body, err := c.MarshalRPCInfo(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	c.callInfos.Set(correlation_id, ClientCallInfo)
	c.Wait() //阻塞等待可以发下一条消息
	_, err = pool.Do("lpush", c.callQueueName, body)
	if err != nil {
		log.Warning("RedisClient Publish: %s", err)
		return err
	}
	return nil
}

/**
消息请求 不需要回复
*/
func (c *redisClient ) CallNR(callInfo rpc.CallInfo) error {
	c.Wait() //阻塞等待可以发下一条消息
	pool := basedb.GetRedisFactory().GetPool(c.redisInfo.RPCUri).Get()
	defer pool.Close()
	var err error

	body, err := c.MarshalRPCInfo(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	c.Wait() //阻塞等待可以发下一条消息
	_, err = pool.Do("lpush", c.callQueueName, body)
	if err != nil {
		log.Warning("Publish: %s", err)
		return err
	}
	return nil
}

func (c *redisClient) on_timeout_handle(done chan error) {
	timeout := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-timeout.C:
			timeout.Reset(time.Second * 1)
			for tuple := range c.callInfos.IterBuffered() {
				var ccInfo = tuple.Val.(*ClientCallInfo)
				if ccInfo.timeout < (time.Now().UnixNano() / 1000000) {
					c.Finish() //完成一个请求
					//从Map中删除
					c.callInfos.Remove(tuple.Key)
					//已经超时了
					resultInfo := &rpcpb.ResultInfo{
						Results:     nil,
						ResultsType: nil,
						Error:      "timeout: This is Call",
					}
					//发送一个超时的消息
					ccInfo.call_chan <- *resultInfo
					//关闭管道
					close(ccInfo.call_chan)
				}
			}
		case <-done:
			timeout.Stop()
			goto LLForEnd

		}
	}
LLForEnd:
}

/**
接收应答信息
*/
func (c *redisClient ) on_response_handle(done chan error) {
	for !c.isClose {
		pool := basedb.GetRedisFactory().GetPool(c.redisInfo.RPCUri).Get()
		result, err := pool.Do("brpop", c.callbackQueueName, 0)//waiting for message accepted
		pool.Close()
		if err == nil && result != nil {
			c.Finish() //完成一个请求
			resultInfo, err := c.UnmarshalResultInfo(result.([]interface{})[1].([]byte))
			if err != nil {
				log.Error("UnmarshalResultInfo faild", err)
			} else {
				correlation_id := resultInfo.Cid
				info, ok := c.callInfos.Get(correlation_id)
				if ok {
					c.callInfos.Remove(correlation_id)
					info.(ClientCallInfo).call_chan <- *resultInfo
					close(info.(ClientCallInfo).call_chan)
				} else {
					//可能客户端已超时了，但服务端处理完还给回调了
					log.Warning("rpc callback no found : [%s]", correlation_id)
				}
			}
		} else if err != nil {
			log.Warning("error %s", err.Error())
		}
	}
	log.Debug("RedisClient finish on_response_handle")
}

func (c *redisClient ) UnmarshalResultInfo(data []byte) (*rpcpb.ResultInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var resultInfo rpcpb.ResultInfo
	err := proto.Unmarshal(data, &resultInfo)
	if err != nil {
		return nil, err
	} else {
		return &resultInfo, err
	}
}

func (c *redisClient ) UnmarshalRPCInfo(data []byte) (*rpcpb.RPCInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var rpcInfo rpcpb.RPCInfo
	err := proto.Unmarshal(data, &rpcInfo)
	if err != nil {
		return nil, err
	} else {
		return &rpcInfo, err
	}
}

// goroutine safe
func (c *redisClient ) MarshalRPCInfo(rpcInfo *rpcpb.RPCInfo) ([]byte, error) {
	//map2:= structs.Map(callInfo)
	b, err := proto.Marshal(rpcInfo)
	return b, err
}
