package baserpc

import (
	"fmt"
	"sync"
	"github.com/streadway/amqp"
	"github.com/golang/protobuf/proto"
	"github.com/dming/lodos/conf"
	"github.com/dming/lodos/rpc/pb"
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/module"
	log "github.com/dming/lodos/log"
	"github.com/dming/lodos/rpc/utils"
	"time"
)

type mqClient struct {
	app module.AppInterface

	callInfos map[string]*ClientCallInfo
	CallChan chan *rpc.CallInfo
	cmutex    sync.Mutex //操作callinfos的锁
	Consumer  *Consumer
	done      chan error

	ticker *time.Ticker
}
/*
type ClientCallInfo struct {
	flag string
	timeout int64 //超时 timeout := time.Now().UnixNano() + ci.expired * 1000000
	chanRet chan *rpc.RetInfo
}*/

func NewMqClient(app module.AppInterface, info *conf.Rabbitmq) (client *mqClient, err error) {
	// Init consumer first
	c, err := NewConsumer(info, info.Uri, info.Exchange, info.ExchangeType, info.ConsumerTag)
	if err != nil {
		//log.Error("AMQPClient connect fail %s", err)
		return
	}
	// 声明回调队列，再次声明的原因是，服务器和客户端可能先后开启，该声明是幂等的，多次声明，但只生效一次
	queue, err := c.channel.QueueDeclare(
		"",    // name of the queue
		false, // durable	持久化
		true,  // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}
	//log.Printf("declared Exchange, declaring Queue %q", queue.Name)
	c.callback_queue = queue.Name //设置callback_queue

	log.Info("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck 自动应答
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}
	client = new(mqClient)
	client.app = app
	client.callInfos = make(map[string]*ClientCallInfo)
	client.Consumer = c
	client.done = make(chan error)
	client.ticker = time.NewTicker(time.Second * 3)
	go client.Run(deliveries, client.done)
	//client.on_timeout_handle(nil) //处理超时请求的协程
	return client, nil
}

func (c *mqClient) Done() (err error) {
	//关闭amqp链接通道
	err = c.Consumer.Shutdown()
	//清理 callinfos 列表
	for key, clinetCallInfo := range c.callInfos {
		if clinetCallInfo != nil {
			//关闭管道
			close(clinetCallInfo.chanRet)
			//从Map中删除
			delete(c.callInfos, key)
		}
	}
	c.callInfos = nil
	return
}

/**
消息请求
*/
func (c *mqClient) Call(ci *rpc.CallInfo) error {
	var err error
	if c.callInfos == nil {
		return fmt.Errorf("MqClient is closed")
	}
	ci.ReplyTo = c.Consumer.callback_queue
	body, err := c.Marshal(ci)
	if err != nil {
		return err
	}

	var flag = ci.Flag
	clientCallInfo := &ClientCallInfo {
		flag: 		flag,
		chanRet: 	ci.ChanRet,
		timeout:    time.Now().UnixNano() + int64(time.Second * 60), //60 * 1000 * 1000, //统一设置成最长超时时间一分钟
	}
	c.callInfos[flag] = clientCallInfo

	if err = c.Consumer.channel.Publish (
		c.Consumer.info.Exchange,   // publish to an exchange
		c.Consumer.info.BindingKey, // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{"reply_to": c.Consumer.callback_queue},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		//log.Warning("Exchange Publish: %s", err)
		log.Warning("Exchange Publish: %s", err)
		return err
	}
	return nil
}


func (c *mqClient) on_timeout_handle() {
	if c.callInfos != nil {
		//处理超时的请求
		for key, clientCallInfo := range c.callInfos {
			if clientCallInfo != nil {
				//var clinetCallInfo = clienCallInfo.(ClientCallInfo)
				if clientCallInfo.timeout < (time.Now().UnixNano() / 1000000) {
					//已经超时了
					retInfo := &rpc.RetInfo{
						Ret: nil,
						Err:  fmt.Errorf("timeout: This is Call"),
					}
					//发送一个超时的消息
					clientCallInfo.chanRet <- retInfo
					//关闭管道
					close(clientCallInfo.chanRet)
					//从Map中删除
					delete(c.callInfos, key)
				}
			}
		}
		//timer.SetTimer(1, c.on_timeout_handle, nil)
	}
}


/**
接收应答信息
*/
func (c *mqClient) Run(deliveries <-chan amqp.Delivery, done chan error) {
	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				deliveries = nil
			} else {
				/*fmt.Printf(
					"MqClient : got %dB on_response_handle delivery: [%v] %q \n",
					len(d.Body),
					d.DeliveryTag,
					d.Body,
				)*/
				d.Ack(false)
				mqRetInfo,err := c.UnmarshalMqRetInfo(d.Body)
				if err != nil {
					log.Error("MqClient : mqRetInfo err is %s \n", err.Error() )
				} else {
					flag := mqRetInfo.Flag

					if clientCallInfo, ok := c.callInfos[flag]; ok {
						retInfo := new(rpc.RetInfo)
						retInfo.Flag = mqRetInfo.Flag
						retInfo.ReplyTo = mqRetInfo.ReplyTo
						if mqRetInfo.Error == "" {
							retInfo.Err = nil
						} else {
							retInfo.Err = fmt.Errorf(mqRetInfo.Error)
						}
						//more
						if len(mqRetInfo.RetsType) == len(mqRetInfo.Rets) {
							retInfo.Ret = make([]interface{}, len(mqRetInfo.RetsType))
							for i, _ := range mqRetInfo.RetsType {
								retInfo.Ret[i], err = argsutils.Bytes2Args(c.app, mqRetInfo.RetsType[i],
									mqRetInfo.Rets[i])
								if err != nil {
									retInfo.Err = err
								}
							}
						}
						clientCallInfo.chanRet <- retInfo

						//删除
						delete(c.callInfos, flag)
					} else {
						fmt.Printf("mqRetInfo.flag of %s not exist \n", mqRetInfo.Flag)
					}
				}
			}
		case <-c.ticker.C:
			c.on_timeout_handle() //execute it every 3 seconds
			break
		case <-done:
			c.Done()
			break
		}
		if deliveries == nil {
			break
		}
	}
}

func (c *mqClient) UnmarshalMqRetInfo(data []byte) (*rpcpb.MqRetInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var mqRetInfo rpcpb.MqRetInfo
	err := proto.Unmarshal(data, &mqRetInfo)
	if err != nil {
		fmt.Printf("UnmarshalMqRetInfo error : %s\n", err)
		return nil, err
	} else {
		return &mqRetInfo, err
	}
}

/*
func (c *mqClient) UnmarshalMqCallInfo(data []byte) (*rpcpb.MqCallInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var mqCallInfo rpcpb.MqCallInfo
	err := proto.Unmarshal(data, &mqCallInfo)
	if err != nil {
		return nil, err
	} else {
		return &mqCallInfo, err
	}

	panic("bug")
}
*/

// goroutine safe
func (c *mqClient) Marshal(ci *rpc.CallInfo) ([]byte, error) {

	mci := new(rpcpb.MqCallInfo)
	mci.Id = ci.Id
	mci.ReplyTo = ci.ReplyTo
	mci.Flag = ci.Flag
	if len(ci.Args) > 0 {
		mci.ArgsType = make([]string, len(ci.Args))
		mci.Args = make([][]byte, len(ci.Args))
		for i, v := range ci.Args {
			var err error
			mci.ArgsType[i], mci.Args[i], err = argsutils.ArgsTypeAnd2Bytes(c.app, v)
			if err != nil {
				return nil, err
			}
		}
	}

	b, err := proto.Marshal(mci)
	return b, err
}


func (c *mqClient) GetDoneChan() (chan error)  {
	return c.done
}
