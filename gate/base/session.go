package basegate

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/dming/lodos/module"
	log "github.com/dming/lodos/log"
	//"sync"
	"reflect"
	"github.com/opentracing/opentracing-go"
	"github.com/dming/lodos/gate"
)

type session struct {
	app       module.AppInterface
	sessionpb   *sessionpb
	//lock *sync.RWMutex
	span opentracing.Span
	judgeGuest func(session gate.Session) bool
}



func NewSession(app module.AppInterface, data []byte) (gate.Session,error) {
	s := &session{
		app : app,
		//lock : new(sync.RWMutex),
	}
	se := &sessionpb{}
	err := proto.Unmarshal(data, se)
	if err != nil {
		return nil,err
	}    // 测试结果
	s.sessionpb = se
	return s, nil
}

func NewSessionByMap(app module.AppInterface, data map[string]interface{}) (gate.Session,error) {
	s:= &session{
		app:app,
		sessionpb:new(sessionpb),
		//lock : new(sync.RWMutex),
	}
	err := s.updateMap(data)
	if err != nil{
		return nil, err
	}
	s.judgeGuest = app.GetJudgeGuest()
	return s, nil
}

func (this *session) GetIP() string {
	return this.sessionpb.GetIP()
}

func (this *session) GetNetwork() string {
	return this.sessionpb.GetNetwork()
}

func (this *session) GetUserid() string {
	return this.sessionpb.GetUserid()
}

func (this *session) GetSessionid() string {
	return this.sessionpb.GetSessionid()
}

func (this *session) GetServerid() string {
	return this.sessionpb.GetServerid()
}

func (this *session) GetSettings() map[string]string {
	return this.sessionpb.GetSettings()
}


func (this *session)SetIP(ip string){
	this.sessionpb.IP=ip
}
func (this *session)SetNetwork(network string){
	this.sessionpb.Network=network
}
func (this *session)SetUserid(userid string){
	this.sessionpb.Userid=userid
}
func (this *session)SetSessionid(sessionid string){
	this.sessionpb.Sessionid=sessionid
}
func (this *session)SetServerid(serverid string){
	this.sessionpb.Serverid=serverid
}
func (this *session)SetSettings(settings map[string]string){
	this.sessionpb.Settings=settings
}

func (this *session) updateMap(settings map[string]interface{}) error {
	var err error = nil
	defer func() {
		if r := recover(); r != nil {
			log.Error("Error on update map, [%settings]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	Userid := settings["Userid"]
	if Userid != nil {
		if result, ok := Userid.(string); ok {
			this.sessionpb.Userid = result
		} else {
			err = fmt.Errorf("UserId can not format to string")
		}
	}
	IP := settings["IP"]
	if IP != nil {
		if result, ok := IP.(string); ok {
			this.sessionpb.IP = result
		} else {
			err = fmt.Errorf("IP can not format to string")
		}
	}
	Network := settings["Network"]
	if Network != nil {
		if result, ok := Network.(string); ok {
			this.sessionpb.Network = result
		} else {
			err = fmt.Errorf("Network can not format to string")
		}
	}
	Sessionid := settings["Sessionid"]
	if Sessionid != nil {
		if result, ok := Sessionid.(string); ok {
			this.sessionpb.Sessionid = result
		} else {
			err = fmt.Errorf("Sessionid can not format to string")
		}
	}
	Serverid := settings["Serverid"]
	if Serverid != nil {
		if result, ok := Serverid.(string); ok {
			this.sessionpb.Serverid = result
		} else {
			err = fmt.Errorf("Serverid can not format to string")
		}
	}
	Settings := settings["Settings"]
	if Settings != nil {
		if result, ok := Settings.(map[string]string); ok {
			this.sessionpb.Settings = result
		} else {
			err = fmt.Errorf("Settings can not format to map[string]string")
		}
	}
	return err
}

func (this *session) update (s gate.Session) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("Error on update, [%this]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	this.sessionpb.Userid = s.GetUserid()

	this.sessionpb.IP = s.GetIP()

	this.sessionpb.Network = s.GetNetwork()

	this.sessionpb.Sessionid = s.GetSessionid()

	this.sessionpb.Serverid = s.GetServerid()

	this.sessionpb.Settings = s.GetSettings()

	return nil
}

func (this *session) Serializable() ([]byte, error){
	data, err := proto.Marshal(this.sessionpb)
	if err != nil {
		return nil, err
	}    // 进行解码
	return data, nil
}

// call Update in gateHandler
func (this *session) Update() (error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("error in update [%this]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	server, err := this.app.GetServerById(this.sessionpb.GetServerid())
	if err != nil {
		return fmt.Errorf("In Update, Service not found id(%this)", this.sessionpb.GetServerid())
	}

	results, err := server.Call("Update", this.sessionpb.GetSessionid())
	if err != nil {
		return err
	}
	if results != nil && len(results) > 0 {
		//成功,重新更新当前Session
		if r, ok := results[0].(gate.Session); ok {
			this.update(r)
		} else {
			return fmt.Errorf("can not convert results[0] to gate.Session")
		}
	}
	return err
}

func (this *session) Bind(Userid string) (error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("error in Bind [%this]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	server, err := this.app.GetServerById(this.sessionpb.GetServerid())
	if err != nil {
		return fmt.Errorf("in Bind, Service not found id(%this)", this.sessionpb.GetServerid())
	}

	results, err := server.Call("Bind",  this.sessionpb.Sessionid, Userid)
	//log.Debug("in Bind, results is : %v", results)
	if err != nil {
		return err
	}
	if results != nil && len(results) > 0 {
		if r, ok := results[0].(gate.Session); ok {
			this.update(r)
		} else {
			return fmt.Errorf("%this can not convert results[0] to Session", reflect.TypeOf(results[0]).String())
		}
	}
	return err
}

func (this *session) UnBind() (error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("error in Bind [%this]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}
	server, err := this.app.GetServerById(this.sessionpb.GetServerid())
	if err != nil {
		return fmt.Errorf("In UnBind ,Service not found id(%this), err is %this", this.sessionpb.GetServerid(), err)
	}

	result, err := server.Call("UnBind", this.sessionpb.Sessionid)
	if err != nil {
		return err
	}
	if result != nil && len(result) > 0 {
		if r, ok := result[0].(gate.Session); ok {
			//绑定成功,重新更新当前Session
			this.update(r)
		} else {
			return fmt.Errorf("can not convert result[0] to this")
		}
	}
	return err
}

func (this *session) Push() (error) {
	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	server, err := this.app.GetServerById(this.sessionpb.Serverid)
	if err != nil {
		return fmt.Errorf("Service not found id(%this)", this.sessionpb.Serverid)
	}

	result, err := server.Call("Push", this.sessionpb.Sessionid, this.sessionpb.Settings)
	if err != nil {
		return err
	}
	if result != nil && len(result) > 0 {
		if r, ok := result[0].(gate.Session); ok {
			//绑定成功,重新更新当前Session
			this.update(r)
		} else {
			return fmt.Errorf("can not convert result[0] to this")
		}
	}
	return err
}

func (this *session) Set(key string, value string) (error) {
	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	if this.sessionpb.Settings == nil {
		this.sessionpb.Settings = map[string]string{}
	}
	//this.lock.Lock()
	this.sessionpb.Settings[key] = value
	//this.lock.Unlock()
	//server,e:=this.app.GetServersById(this.Serverid)
	//if e!=nil{
	//	err=fmt.Sprintf("Service not found id(%this)",this.Serverid)
	//	return
	//}
	//result,err:=server.Call("Set",this.Sessionid,key,value)
	//if err==""{
	//	if result!=nil{
	//		//绑定成功,重新更新当前Session
	//		this.update(result.(map[string]interface {}))
	//	}
	//}
	return nil
}

func (this *session) SetPush(key string, value string) error {
	if this.app == nil {
		return fmt.Errorf("Module.App is nil")
	}
	if this.sessionpb.Settings == nil {
		this.sessionpb.Settings = map[string]string{}
	}
	this.sessionpb.Settings[key] = value
	return this.Push()
}

func (this *session) Get(key string) (result string) {
	if this.sessionpb.Settings == nil {
		return
	}
	//this.lock.RLock()
	result = this.sessionpb.Settings[key]
	//this.lock.RUnlock()
	return
}

func (this *session) Remove(key string) (error) {
	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	if this.sessionpb.Settings == nil {
		this.sessionpb.Settings=map[string]string{}
	}
	//this.lock.Lock()
	delete(this.sessionpb.Settings, key)
	//this.lock.Unlock()
	//server,e:=this.app.GetServersById(this.Serverid)
	//if e!=nil{
	//	err=fmt.Sprintf("Service not found id(%this)",this.Serverid)
	//	return
	//}
	//result,err:=server.Call("Remove",this.Sessionid,key)
	//if err==""{
	//	if result!=nil{
	//		//绑定成功,重新更新当前Session
	//		this.update(result.(map[string]interface {}))
	//	}
	//}
	return nil
}

func (this *session) Send(topic string, body []byte) (error) {
	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}
	server, e := this.app.GetServerById(this.sessionpb.Serverid)
	if e != nil {
		return fmt.Errorf("Service not found id(%this)", this.sessionpb.Serverid)
	}

	_, err := server.Call("Send", this.sessionpb.Sessionid, topic, body)
	return err
}

func (this *session) SendBatch(Sessionids string, topic string, body []byte) (int64, error) {
	if this.app == nil {
		return 0, fmt.Errorf("Module.App is nil")
	}
	server, e := this.app.GetServerById(this.sessionpb.Serverid)
	if e != nil {
		return 0, fmt.Errorf("Service not found id(%s)", this.sessionpb.Serverid)
	}
	count, err := server.Call("SendBatch", Sessionids, topic, body)
	if err != nil {
		return 0, err
	}
	return count[0].(int64), err
}

func (this *session) SendNR(topic string, body []byte) (error) {
	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	server, err := this.app.GetServerById(this.sessionpb.Serverid)
	if err != nil {
		return fmt.Errorf("Service not found id(%this)", this.sessionpb.Serverid)
	}

	err = server.CallNR("Send", this.sessionpb.Sessionid, topic, body)
	return err
}

func (this *session) Close() (error) {
	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	server, err := this.app.GetServerById(this.sessionpb.Serverid)
	if err != nil {
		return  fmt.Errorf("Service not found id(%this)", this.sessionpb.Serverid)
	}

	_, err = server.GetClient().Call("Close", 5, this.sessionpb.Sessionid)
	return err
}

/**
每次rpc调用都拷贝一份新的Session进行传输
*/
func (this *session) Clone() gate.Session {
	s := &session{
		app:  this.app,
		span: this.Span(),
	}
	se := &sessionpb{
		IP:        this.sessionpb.IP,
		Network:   this.sessionpb.Network,
		Userid:    this.sessionpb.Userid,
		Sessionid: this.sessionpb.Sessionid,
		Serverid:  this.sessionpb.Serverid,
		Settings:  this.sessionpb.Settings,
	}
	//这个要换成本次RPC调用的新Span
	se.Carrier = this.inject()

	s.sessionpb = se
	return s
}

func (this *session) IsConnect(userId string) (bool, error) {
	if this.app == nil {
		return false, fmt.Errorf("Module.App is nil")
	}
	server, e := this.app.GetServerById(this.sessionpb.Serverid)
	if e != nil {
		return false, fmt.Errorf("Service not found id(%s)", this.sessionpb.Serverid)
	}
	result, err := server.Call("IsConnect", this.sessionpb.Sessionid, userId)
	return result[0].(bool), err
}


func (this *session) inject() map[string]string {
	if this.app.GetTracer() == nil {
		return nil
	}
	if this.Span() == nil {
		return nil
	}
	carrier := &opentracing.TextMapCarrier{}
	err := this.app.GetTracer().Inject(
		this.Span().Context(),
		opentracing.TextMap,
		carrier)
	if err != nil {
		log.Warning("session.session.Carrier Inject Fail", err.Error())
		return nil
	} else {
		m := map[string]string{}
		carrier.ForeachKey(func(key, val string) error {
			m[key] = val
			return nil
		})
		return m
	}
}
func (this *session) extract(gCarrier map[string]string) (opentracing.SpanContext, error) {
	carrier := &opentracing.TextMapCarrier{}
	for v, k := range gCarrier {
		carrier.Set(v, k)
	}
	return this.app.GetTracer().Extract(opentracing.TextMap, carrier)
}
func (this *session) LoadSpan(operationName string) opentracing.Span {
	if this.app.GetTracer() == nil {
		return nil
	}
	if this.span == nil {
		if this.sessionpb.Carrier != nil {
			//从已有记录恢复
			clientContext, err := this.extract(this.sessionpb.Carrier)
			if err == nil {
				this.span = this.app.GetTracer().StartSpan(
					operationName, opentracing.ChildOf(clientContext))
			} else {
				log.Warning("session.session.Carrier Extract Fail", err.Error())
			}
		}
	}
	return this.span
}

func (this *session) CreateRootSpan(operationName string) opentracing.Span {
	if this.app.GetTracer() == nil {
		return nil
	}
	this.span = this.app.GetTracer().StartSpan(operationName)
	this.sessionpb.Carrier = this.inject()
	return this.span
}

func (this *session) Span() opentracing.Span {
	return this.span
}

func (this *session) TracingCarrier() map[string]string {
	return this.sessionpb.Carrier
}

func (this *session) TracingId() string {
	if this.TracingCarrier() != nil {
		if tid, ok := this.TracingCarrier()["ot-tracer-traceid"]; ok {
			return tid
		}
	}
	return ""
}

/**
从Session的 Span继承一个新的Span
*/
func (this *session) ExtractSpan(operationName string) opentracing.Span {
	if this.app.GetTracer() == nil {
		return nil
	}
	if this.Span() != nil {
		span := this.app.GetTracer().StartSpan(operationName, opentracing.ChildOf(this.Span().Context()))
		return span
	}
	return nil
}

//是否是访客(未登录) ,默认判断规则为 userId==""代表访客
func (this *session) IsGuest() bool {
	if this.judgeGuest != nil {
		return this.judgeGuest(this)
	}
	if this.GetUserid() == "" {
		return true
	} else {
		return false
	}
}

//设置自动的访客判断函数,记得一定要在全局的时候设置这个值,以免部分模块因为未设置这个判断函数造成错误的判断
func (this *session) SetJudgeGuest(judgeGuest func(session gate.Session) bool) {
	this.judgeGuest = judgeGuest
}