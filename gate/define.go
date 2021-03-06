// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package gate

import (
	"github.com/opentracing/opentracing-go"
	"github.com/dming/lodos/network"
)

type GateInterface interface {
	GetMinStorageHeartbeat() int64
	GetGateHandler() GateHandler
	GetAgentLearner() AgentLearner
	GetSessionLearner() SessionLearner
	GetStorageHandler() StorageHandler
	GetTracingHandler() TracingHandler
	NewSession(data []byte) (Session, error)
	NewSessionByMap(data map[string]interface{}) (Session, error)
}

/**
net代理服务 处理器
*/
type GateHandler interface {
	Bind(Sessionid string, Userid string) (result Session, err error)                      //Bind the session with the the Userid.
	UnBind(Sessionid string) (result Session, err error)                                   //UnBind the session with the the Userid.
	Set(Sessionid string, key string, value string) (result Session, err error)            //Set values (one or many) for the session.
	Remove(Sessionid string, key string) (result Session, err error)                       //Remove value from the session.
	Push(Sessionid string, Settings map[string]string) (result Session, err error) //推送信息给Session
	Send(Sessionid string, topic string, body []byte) (err error)                          //Send message to the session.
	SendBatch(Sessionids string, topic string, body []byte) (int64, error)                //批量发送
	BroadCast(topic string, body []byte) (int64, error)
	Update(Sessionid string) (result Session, err error) //更新整个Session 通常是其他模块拉取最新数据
	//查询某一个userId是否连接中，这里只是查询这一个网关里面是否有userId客户端连接，如果有多个网关就需要遍历了
	IsConnect(Sessionid string, Userid string) (result bool, err error)
	Close(Sessionid string) (err error)                  //主动关闭连接
	OnDestroy()
}

type Session interface {
	GetIP() string
	GetNetwork() string
	GetUserid() string
	GetSessionid() string
	GetServerid() string
	GetSettings() map[string]string
	GetCarrier() map[string]string
	GetToken() string
	SetIP(ip string)
	SetNetwork(network string)
	SetUserid(userid string)
	SetSessionid(sessionid string)
	SetServerid(serverid string)
	SetSettings(settings map[string]string)
	SetCarrier(carrier map[string]string)
	SetToken(token string)
	Serializable()([]byte, error)
	Update() (error)
	Bind(Userid string) (error)
	UnBind() (error)
	Push() (error)
	Set(key string, value string) (error)
	SetPush(key string, value string) (error) //设置值以后立即推送到gate网关
	Get(key string) (result string)
	Remove(key string) (error)
	Send(topic string, body []byte) (error)
	SendNR(topic string, body []byte) (error)
	Close() (error)
	Clone() Session

	//查询某一个userId是否连接中，这里只是查询这一个网关里面是否有userId客户端连接，如果有多个网关就需要遍历了
	IsConnect(userId string) (result bool, err error)
	//是否是访客(未登录) ,默认判断规则为 userId==""代表访客
	IsGuest() bool
	//设置自动的访客判断函数,记得一定要在全局的时候设置这个值,以免部分模块因为未设置这个判断函数造成错误的判断
	SetJudgeGuest(judgeGuest func(session Session) bool)
	/**
	通过Carrier数据构造本次rpc调用的tracing Span,如果没有就创建一个新的
	*/
	CreateRootSpan(openrationName string) opentracing.Span
	/**
	通过Carrier数据构造本次rpc调用的tracing Span,如果没有就返回nil
	*/
	LoadSpan(openrationName string) opentracing.Span
	/**
	获取本次rpc调用的tracing Span
	*/
	Span() opentracing.Span
	/**
	从Session的 Span继承一个新的Span
	*/
	ExtractSpan(operationName string) opentracing.Span

	/**
	获取Tracing的Carrier 可能为nil
	*/
	TracingCarrier() map[string]string
	TracingId() string
}

/**
Session信息持久化
*/
type StorageHandler interface {
	/**
	存储用户的Session信息
	Session Bind Userid以后每次设置 settings都会调用一次Storage
	*/
	Storage(Userid string, session Session) (err error)
	/**
	强制删除Session信息
	*/
	Delete(Userid string) (err error)
	/**
	获取用户Session信息
	Bind Userid时会调用Query获取最新信息
	*/
	Query(Userid string) (data []byte, err error)
	/**
	用户心跳,一般用户在线时1s发送一次
	可以用来延长Session信息过期时间
	*/
	Heartbeat(Userid string)
}

type TracingHandler interface {
	/**
	是否需要对本次客户端请求进行跟踪
	*/
	OnRequestTracing(session Session, topic string, msg []byte) bool
}

type AgentLearner interface {
	Connect(a Agent)    //当连接建立  并且MQTT协议握手成功
	DisConnect(a Agent) //当连接关闭	或者客户端主动发送MQTT DisConnect命令
}

type SessionLearner interface {
	Connect(a Session)    //当连接建立  并且MQTT协议握手成功
	DisConnect(a Session) //当连接关闭	或者客户端主动发送MQTT DisConnect命令
}

type Agent interface {
	OnInit(gate GateInterface, conn network.Conn) error
	WriteMsg(topic string, body []byte) error
	Run() error
	Close()
	OnClose() error
	Destroy()
	RevNum() int64
	SendNum() int64
	IsClosed() bool
	GetSession() Session
}
