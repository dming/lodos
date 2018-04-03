package baserpc

import (
	"github.com/dming/lodos/rpc"
	"github.com/dming/lodos/rpc/pb"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
	"fmt"
	"runtime"
	"github.com/dming/lodos/log"
	"github.com/dming/lodos/conf"
)

type amqpServer struct {
	call_chan chan rpc.CallInfo
	rabbitAgent *RabbitAgent
	done chan error
}


func NewAMQPServer(info *conf.Rabbitmq, call_chan chan rpc.CallInfo) (*amqpServer, error) {
	agent, err := NewRabbitAgent(info, TypeServer)
	if err != nil {
		return nil, fmt.Errorf("rabbit agent: %s", err.Error());
	}
	server := new(amqpServer)
	server.call_chan = call_chan
	server.rabbitAgent = agent
	server.done = make(chan error)
	go server.on_requese_handle(agent.ReadMsg(), server.done)
	return server, nil
}

func (s *amqpServer) StopConsume() error {
	return nil
}

//close
func (s *amqpServer) Shutdown() error {
	return s.rabbitAgent.Shutdown()
}

//callback
func (s *amqpServer) Callback(callInfo rpc.CallInfo) error {
	body, _ := s.MarshalResult(callInfo.Result)
	return s.response(callInfo.Props, body)
}

//response the information
func (s *amqpServer) response(props map[string]interface{}, body []byte) error {
	return s.rabbitAgent.ServerPublish(props["reply_to"].(string), body)
}

func (s *amqpServer) on_requese_handle(deliveries <-chan amqp.Delivery, done chan error) {
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

	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				deliveries = nil
			} else {
				d.Ack(false)
				rpcInfo, err := s.UnmarshalRpcInfo(d.Body)
				if (err == nil) {
					callInfo := &rpc.CallInfo{
						RpcInfo: *rpcInfo,
					}
					callInfo.Props = map[string]interface{}{
						"reply_to": callInfo.RpcInfo.ReplyTo,
					}

					callInfo.Agent = s
					s.call_chan <- *callInfo
				} else {
					fmt.Println("error", err)
				}
			}
		case <-done:
			goto LForEnd
		}
		if deliveries == nil {
			goto LForEnd
		}
	}
LForEnd:
}


func (s *amqpServer) MarshalResult(resultInfo rpcpb.ResultInfo) ([]byte, error) {
	b, err := proto.Marshal(&resultInfo)
	return b, err
}

//解码数据
func (s *amqpServer) UnmarshalRpcInfo(data []byte) (*rpcpb.RPCInfo, error) {
	var rpcInfo rpcpb.RPCInfo
	err := proto.Unmarshal(data, &rpcInfo)
	if err != nil {
		return nil, err
	} else {
		return &rpcInfo, err
	}
}