package baserpc

import (
	"github.com/dming/lodos/rpc/pb"
	"time"
	"github.com/dming/lodos/log"
	"github.com/dming/lodos/utils"
	"fmt"
	"github.com/dming/lodos/rpc"
)

type localClient struct {
	//callInfos map[string]*ClientCallInfo
	callInfos    *utils.BeeMap
	//cmutex       sync.Mutex //操作callinfos的锁
	local_server rpc.LocalServer
	result_chan  chan rpcpb.ResultInfo // chan for result return
	done         chan error
	isClose      bool
	//timeout_done chan error
}

// 这里是暴露整个Local server还是只暴露local server的mq_chan呢
// 因为是本地，暂时暴露整个local server也是OK的
func NewLocalClient(server rpc.LocalServer) (*localClient, error) {
	client := new(localClient)
	//client.callInfos=make(map[string]*ClientCallInfo)
	client.callInfos = utils.NewBeeMap()
	client.local_server = server
	client.done = make(chan error)
	//client.timeout_done = make(chan error)
	client.isClose = false
	client.result_chan = make(chan rpcpb.ResultInfo, 1)
	go client.on_response_handle(client.result_chan, client.done)
	//go client.on_timeout_handle(client.timeout_done) //处理超时请求的协程
	return client, nil
}

func (c *localClient) Done() error {
	//关闭消息回复通道
	c.isClose = true
	c.done <- nil
	close(c.result_chan)
	//清理 callInfos 列表
	for key, clientCallInfo := range c.callInfos.Items() {
		if clientCallInfo != nil {
			//关闭管道
			close(clientCallInfo.(ClientCallInfo).call_chan)
			//从Map中删除
			c.callInfos.Delete(key)
		}
	}
	return nil
}

/**
消息请求, and function Call() is sync
*/
func (c *localClient) Call(callInfo rpc.CallInfo, callback_chan chan rpcpb.ResultInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()

	if c.isClose {
		return fmt.Errorf("MQClient is isClose")
	}

	var correlation_id = callInfo.RpcInfo.Cid

	ClientCallInfo := &ClientCallInfo{
		correlation_id: correlation_id,
		call_chan:           callback_chan,
		timeout:        callInfo.RpcInfo.Expired,
	}
	c.callInfos.Set(correlation_id, *ClientCallInfo)
	callInfo.Agent = c.local_server
	callInfo.Props = map[string]interface{}{
		"reply_to": c.result_chan,
	}
	//发送消息
	c.local_server.WriteToRpcServer(callInfo)

	return nil
}

/**
消息请求 不需要回复
*/
func (c *localClient) CallNR(callInfo rpc.CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()
	if c.isClose {
		return fmt.Errorf("MQClient is isClose")
	}
	//发送消息
	c.local_server.WriteToRpcServer(callInfo)

	return nil
}

//func (c *localClient) on_timeout_handle(done chan error) {
//	timeout := time.NewTimer(time.Second * 1)
//	for {
//		select {
//		case <-timeout.C:
//			timeout.Reset(time.Second * 1)
//			for key, ClientCallInfo := range c.callInfos.Items() {
//				if ClientCallInfo != nil {
//					var ClientCallInfo = ClientCallInfo.(ClientCallInfo)
//					if ClientCallInfo.timeout < (time.Now().UnixNano() / 1000000) {
//						//从Map中删除
//						c.callInfos.Delete(key)
//						//已经超时了
//						resultInfo := &rpcpb.ResultInfo{
//							Result:     nil,
//							Error:      "timeout: This is Call",
//							ResultType: argsutil.NULL,
//						}
//						//发送一个超时的消息
//						ClientCallInfo.call_chan <- *resultInfo
//						//关闭管道
//						close(ClientCallInfo.call_chan)
//					}
//
//				}
//			}
//		case <-done:
//			timeout.Stop()
//			goto LLForEnd
//
//		}
//	}
//LLForEnd:
//}

func (c *localClient) on_timeout_handle(args interface{}) {
	//处理超时的请求
	for key, clientCallInfo := range c.callInfos.Items() {
		if clientCallInfo != nil {
			var clientCallInfo = clientCallInfo.(ClientCallInfo)
			if clientCallInfo.timeout < (time.Now().UnixNano() / 1000000) {
				//从Map中删除
				c.callInfos.Delete(key)
				//已经超时了
				resultInfo := rpcpb.NewResultInfo(
					"",
					ServerExpired,
					fmt.Sprintf("timeout: This CallInfo : %s timeout!", clientCallInfo.correlation_id),
					nil, nil)
				//发送一个超时的消息
				clientCallInfo.call_chan <- *resultInfo
				//关闭管道
				close(clientCallInfo.call_chan)
			}
		}
	}
}

/**
接收应答信息
*/
func (c *localClient) on_response_handle(deliveries <-chan rpcpb.ResultInfo, done chan error) {
	timeout := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-done:
			goto ForEnd
		case resultInfo, ok := <-deliveries:
			if !ok {
				deliveries = nil
			} else {
				correlation_id := resultInfo.Cid
				if clientCallInfo := c.callInfos.Get(correlation_id); clientCallInfo != nil {
					//删除
					c.callInfos.Delete(correlation_id)
					clientCallInfo.(ClientCallInfo).call_chan <- resultInfo
					close(clientCallInfo.(ClientCallInfo).call_chan)
				} else {
					//可能客户端已超时了，但服务端处理完还给回调了
					log.Warning("rpc callback no found : [%s]", correlation_id)
				}
			}
		case <-timeout.C: // every second check timeout
			timeout.Reset(time.Second * 1)
			c.on_timeout_handle(nil)
		}

		if deliveries == nil {
			goto ForEnd
		}
	}
ForEnd:
	timeout.Stop()
}
