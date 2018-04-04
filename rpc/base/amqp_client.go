package baserpc

import (
	"github.com/dming/lodos/rpc/pb"
	"github.com/golang/protobuf/proto"
	"github.com/dming/lodos/utils"
	"github.com/dming/lodos/conf"
	"fmt"
	"time"
	"github.com/streadway/amqp"
	"github.com/dming/lodos/log"
	"github.com/dming/lodos/rpc"
)

type amqpClient struct {
	//callInfos map[string]*ClientCallInfo
	callInfos   *utils.BeeMap
	//cmutex      sync.Mutex //操作callinfos的锁
	rabbitAgent *RabbitAgent
	done        chan error
	isClose bool
}

func NewAMQPClient(info *conf.Rabbitmq) (client *amqpClient, err error) {
	agent, err := NewRabbitAgent(info, TypeClient)
	if err != nil {
		return nil, fmt.Errorf("rabbit agent: %s", err.Error())
	}
	client = new(amqpClient)
	client.callInfos = utils.NewBeeMap()
	client.rabbitAgent = agent
	client.done = make(chan error)
	client.isClose = false
	go client.on_response_handle(agent.ReadMsg(), client.done)
	return client, nil
}

func (c *amqpClient) Done() (err error) {
	c.isClose = true
	//关闭amqp链接通道
	err = c.rabbitAgent.Shutdown()
	c.done <- nil
	close(c.done)
	//c.send_done<-nil
	//c.done<-nil
	//清理 callInfos 列表
	for key, clientCallInfo := range c.callInfos.Items() {
		if clientCallInfo != nil {
			//关闭管道
			close(clientCallInfo.(ClientCallInfo).call_chan)
			//从Map中删除
			c.callInfos.Delete(key)
		}
	}
	c.callInfos = nil
	return
}

/**
消息请求
*/
func (c *amqpClient) Call(callInfo rpc.CallInfo, callback_chan chan rpcpb.ResultInfo) error {
	//var err error
	if c.isClose {
		return fmt.Errorf("amqpClient is isClose")
	}
	callback_queue, err := c.rabbitAgent.CallbackQueue()
	if err != nil {
		return err
	}
	callInfo.RpcInfo.ReplyTo = callback_queue
	var correlation_id = callInfo.RpcInfo.Cid

	clientCallInfo := &ClientCallInfo{
		correlation_id: correlation_id,
		call_chan:           callback_chan,
		timeout:        callInfo.RpcInfo.Expired,
	}
	c.callInfos.Set(correlation_id, *clientCallInfo)
	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	return c.rabbitAgent.ClientPublish(body)
}

/**
消息请求 不需要回复
*/
func (c *amqpClient) CallNR(callInfo rpc.CallInfo) error {
	if c.isClose {
		return fmt.Errorf("amqpClient is isClose")
	}
	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	return c.rabbitAgent.ClientPublish(body)
}

func (c *amqpClient) on_timeout_handle(args interface{}) {
	if c.callInfos != nil {
		//处理超时的请求
		for key, info := range c.callInfos.Items() {
			if info != nil {
				var ccInfo = info.(ClientCallInfo)
				if ccInfo.timeout < (time.Now().UnixNano() / 1000000) {
					//从Map中删除
					c.callInfos.Delete(key)
					//已经超时了
					resultInfo := &rpcpb.ResultInfo{
						Cid: ccInfo.correlation_id,
						Results:     nil,
						Error:      fmt.Sprintf("timeout: This CallInfo : %s timeout!", ccInfo.correlation_id),
						ResultsType: nil,
					}
					//发送一个超时的消息
					ccInfo.call_chan <- *resultInfo
					//关闭管道
					close(ccInfo.call_chan)
				}
			}
		}
	}
}

/**
接收应答信息
*/
func (c *amqpClient) on_response_handle(deliveries <-chan amqp.Delivery, done chan error) {
	timeout := time.NewTimer(time.Second * 1)
	defer timeout.Stop()

	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				deliveries = nil
			} else {
				d.Ack(false)
				resultInfo, err := c.UnmarshalResultInfo(d.Body)
				if err != nil {
					log.Error("UnmarshalResultInfo faild", err)
				} else {
					correlation_id := resultInfo.Cid
					if ccInfo := c.callInfos.Get(correlation_id); ccInfo != nil {
						//删除
						c.callInfos.Delete(correlation_id)
						ccInfo.(ClientCallInfo).call_chan <- *resultInfo
						close(ccInfo.(ClientCallInfo).call_chan)
					} else {
						//可能客户端已超时了，但服务端处理完还给回调了
						log.Warning("rpc callback no found : [%s]", correlation_id)
					}
				}
			}
		case <-timeout.C:
			timeout.Reset(time.Second * 1)
			c.on_timeout_handle(nil)
		case <-done:
			goto LForEnd
		}
		if deliveries == nil {
			goto LForEnd
		}
	}
LForEnd:
}

func (c *amqpClient) UnmarshalResultInfo(data []byte) (*rpcpb.ResultInfo, error) {
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

func (c *amqpClient) UnmarshalRPCInfo(data []byte) (*rpcpb.RPCInfo, error) {
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
func (c *amqpClient) Marshal(rpcInfo *rpcpb.RPCInfo) ([]byte, error) {
	//map2:= structs.Map(callInfo)
	b, err := proto.Marshal(rpcInfo)
	return b, err
}

