package baserpc

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/log"
	"github.com/dming/lodos/rpc/pb"
	"github.com/dming/lodos/utils"
	"sync"
	"time"
	"github.com/dming/lodos/rpc"
)

type redisClient struct {
	//callinfos map[string]*ClientCallInfo
	callinfos         utils.ConcurrentMap
	cmutex            sync.Mutex //操作callinfos的锁
	info              *conf.Redis
	queueName         string
	callbackqueueName string
	done              chan error
	timeout_done      chan error
	pool              redis.Conn
	closed            bool
	ch                chan int //控制一个RPC 客户端可以同时等待的最大消息量，
	// 如果等待的请求过多代表rpc server请求压力大，
	// 就不要在往消息队列发消息了,客户端先阻塞

}

func createQueueName() string {
	//return "callbackqueueName"
	return fmt.Sprintf("callbackqueueName:%d", time.Now().Nanosecond())
}
func NewRedisClient(info *conf.Redis) (client  *redisClient , err error) {
	client = new(redisClient)
	client.callinfos = utils.New()
	client.info = info
	client.callbackqueueName = createQueueName()
	client.queueName = info.Queue
	client.done = make(chan error)
	client.timeout_done = make(chan error)
	client.closed = false
	client.ch = make(chan int, 100)
	pool := utils.GetRedisFactory().GetPool(info.Uri).Get()
	defer pool.Close()
	//_, errs:=pool.Do("EXPIRE", client.callbackqueueName, 60)
	//if errs != nil {
	//	log.Warning(errs.Error())
	//}
	go client.on_response_handle(client.done)
	go client.on_timeout_handle(client.timeout_done) //处理超时请求的协程
	return client, nil
	//log.Printf("shutting down")
	//
	//if err := c.Shutdown(); err != nil {
	//	log.Fatalf("error during shutdown: %s", err)
	//}
}
func (this  *redisClient ) Wait() error {
	// 如果ch满了则会处于阻塞，从而达到限制最大协程的功能
	//this.ch <- 1
	return nil
}
func (this  *redisClient ) Finish() {
	// 完成则从ch推出数据
	//<-this.ch
}
func (c  *redisClient ) Done() (err error) {
	pool := utils.GetRedisFactory().GetPool(c.info.Uri).Get()
	defer pool.Close()
	//删除临时通道
	pool.Do("DEL", c.callbackqueueName)
	c.closed = true
	c.timeout_done <- nil
	//err = c.psc.Close()
	//清理 callinfos 列表
	if c.pool != nil {
		c.pool.Close()
	}
	for tuple := range c.callinfos.Iter() {
		//关闭管道
		close(tuple.Val.(*ClientCallInfo).call_chan)
		//从Map中删除
		c.callinfos.Remove(tuple.Key)
	}
	return
}

/**
消息请求
*/
func (c  *redisClient ) Call(callInfo rpc.CallInfo, callback chan rpcpb.ResultInfo) error {
	pool := utils.GetRedisFactory().GetPool(c.info.Uri).Get()
	defer pool.Close()
	var err error
	if c.closed {
		return fmt.Errorf("RedisClient is closed")
	}
	callInfo.RpcInfo.ReplyTo = c.callbackqueueName
	var correlation_id = callInfo.RpcInfo.Cid

	ClientCallInfo := &ClientCallInfo{
		correlation_id: correlation_id,
		call_chan:           callback,
		timeout:        callInfo.RpcInfo.Expired,
	}

	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	c.callinfos.Set(correlation_id, ClientCallInfo)
	c.Wait() //阻塞等待可以发下一条消息
	_, err = pool.Do("lpush", c.queueName, body)
	if err != nil {
		log.Warning("Publish: %s", err)
		return err
	}
	return nil
}

/**
消息请求 不需要回复
*/
func (c  *redisClient ) CallNR(callInfo rpc.CallInfo) error {
	c.Wait() //阻塞等待可以发下一条消息
	pool := utils.GetRedisFactory().GetPool(c.info.Uri).Get()
	defer pool.Close()
	var err error

	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	_, err = pool.Do("lpush", c.queueName, body)
	if err != nil {
		log.Warning("Publish: %s", err)
		return err
	}
	return nil
}

func (c  *redisClient ) on_timeout_handle(done chan error) {
	timeout := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-timeout.C:
			timeout.Reset(time.Second * 1)
			for tuple := range c.callinfos.Iter() {
				var ClientCallInfo = tuple.Val.(*ClientCallInfo)
				if ClientCallInfo.timeout < (time.Now().UnixNano() / 1000000) {
					c.Finish() //完成一个请求
					//从Map中删除
					c.callinfos.Remove(tuple.Key)
					//已经超时了
					resultInfo := &rpcpb.ResultInfo{
						Results:     nil,
						ResultsType: nil,
						Error:      "timeout: This is Call",
					}
					//发送一个超时的消息
					ClientCallInfo.call_chan <- *resultInfo
					//关闭管道
					close(ClientCallInfo.call_chan)
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
func (c  *redisClient ) on_response_handle(done chan error) {
	for !c.closed {
		c.pool = utils.GetRedisFactory().GetPool(c.info.Uri).Get()
		result, err := c.pool.Do("brpop", c.callbackqueueName, 0)
		c.pool.Close()
		if err == nil && result != nil {
			c.Finish() //完成一个请求
			resultInfo, err := c.UnmarshalResult(result.([]interface{})[1].([]byte))
			if err != nil {
				log.Error("Unmarshal faild", err)
			} else {
				correlation_id := resultInfo.Cid
				info, ok := c.callinfos.Get(correlation_id)
				if ok {
					c.callinfos.Remove(correlation_id)
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
	log.Debug("finish on_response_handle")
}

func (c  *redisClient ) UnmarshalResult(data []byte) (*rpcpb.ResultInfo, error) {
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

func (c  *redisClient ) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
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
func (c  *redisClient ) Marshal(rpcInfo *rpcpb.RPCInfo) ([]byte, error) {
	//map2:= structs.Map(callInfo)
	b, err := proto.Marshal(rpcInfo)
	return b, err
}
