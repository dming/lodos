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
	Sessionpb   *Sessionpb
	//lock *sync.RWMutex
	span opentracing.Span
	judgeGuest func(session gate.Session) bool
}



func NewSession(app module.AppInterface, data []byte) (gate.Session, error) {
	s := &session{
		app : app,
		//lock : new(sync.RWMutex),
	}
	se := &Sessionpb{}
	err := proto.Unmarshal(data, se)
	if err != nil {
		return nil,err
	}    // 测试结果
	s.Sessionpb = se
	s.judgeGuest = app.GetJudgeGuest()
	return s, nil
}

func NewSessionByMap(app module.AppInterface, data map[string]interface{}) (gate.Session,error) {
	s:= &session{
		app:app,
		Sessionpb:new(Sessionpb),
		//lock : new(sync.RWMutex),
	}
	err := s.updateMap(data)
	if err != nil{
		return nil, err
	}
	s.judgeGuest = app.GetJudgeGuest()
	return s, nil
}

func (this *session) GetApp() module.AppInterface {
	return this.app
}

func (this *session) GetIP() string {
	return this.Sessionpb.GetIP()
}

func (this *session) GetNetwork() string {
	return this.Sessionpb.GetNetwork()
}

func (this *session) GetUserid() string {
	return this.Sessionpb.GetUserid()
}

func (this *session) GetSessionid() string {
	return this.Sessionpb.GetSessionid()
}

func (this *session) GetServerid() string {
	return this.Sessionpb.GetServerid()
}

func (this *session) GetSettings() map[string]string {
	return this.Sessionpb.GetSettings()
}

func (this *session) GetCarrier() map[string]string {
	return this.Sessionpb.GetCarrier()
}

func (this *session) GetToken() string {
	return this.Sessionpb.GetToken()
}

func (this *session)SetIP(ip string){
	this.Sessionpb.IP=ip
}
func (this *session)SetNetwork(network string){
	this.Sessionpb.Network=network
}
func (this *session)SetUserid(userid string){
	this.Sessionpb.Userid=userid
}
func (this *session)SetSessionid(sessionid string){
	this.Sessionpb.Sessionid=sessionid
}
func (this *session)SetServerid(serverid string){
	this.Sessionpb.Serverid=serverid
}
func (this *session)SetSettings(settings map[string]string){
	this.Sessionpb.Settings=settings
}
func (this *session) SetCarrier(carrier map[string]string){
	this.Sessionpb.Carrier = carrier
}
func (this *session) SetToken(token string){
	this.Sessionpb.Token = token
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
			this.Sessionpb.Userid = result
		} else {
			err = fmt.Errorf("UserId can not format to string")
		}
	}
	IP := settings["IP"]
	if IP != nil {
		if result, ok := IP.(string); ok {
			this.Sessionpb.IP = result
		} else {
			err = fmt.Errorf("IP can not format to string")
		}
	}
	Network := settings["Network"]
	if Network != nil {
		if result, ok := Network.(string); ok {
			this.Sessionpb.Network = result
		} else {
			err = fmt.Errorf("Network can not format to string")
		}
	}
	Sessionid := settings["Sessionid"]
	if Sessionid != nil {
		if result, ok := Sessionid.(string); ok {
			this.Sessionpb.Sessionid = result
		} else {
			err = fmt.Errorf("Sessionid can not format to string")
		}
	}
	Serverid := settings["Serverid"]
	if Serverid != nil {
		if result, ok := Serverid.(string); ok {
			this.Sessionpb.Serverid = result
		} else {
			err = fmt.Errorf("Serverid can not format to string")
		}
	}
	Settings := settings["Settings"]
	if Settings != nil {
		if result, ok := Settings.(map[string]string); ok {
			this.Sessionpb.Settings = result
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

	this.Sessionpb.Userid = s.GetUserid()

	this.Sessionpb.IP = s.GetIP()

	this.Sessionpb.Network = s.GetNetwork()

	this.Sessionpb.Sessionid = s.GetSessionid()

	this.Sessionpb.Serverid = s.GetServerid()

	this.Sessionpb.Settings = s.GetSettings()

	return nil
}

func (this *session) Serializable() ([]byte, error){
	data, err := proto.Marshal(this.Sessionpb)
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

	server, err := this.app.GetServerById(this.Sessionpb.GetServerid())
	if err != nil {
		return fmt.Errorf("In Update, Service not found id(%this)", this.Sessionpb.GetServerid())
	}

	results, err := server.Call("Update", this.Sessionpb.GetSessionid())
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

	server, err := this.app.GetServerById(this.Sessionpb.GetServerid())
	if err != nil {
		return fmt.Errorf("in Bind, Service not found id(%this)", this.Sessionpb.GetServerid())
	}

	results, err := server.Call("Bind",  this.Sessionpb.Sessionid, Userid)
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
	server, err := this.app.GetServerById(this.Sessionpb.GetServerid())
	if err != nil {
		return fmt.Errorf("In UnBind ,Service not found id(%this), err is %this", this.Sessionpb.GetServerid(), err)
	}

	result, err := server.Call("UnBind", this.Sessionpb.Sessionid)
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

	server, err := this.app.GetServerById(this.Sessionpb.Serverid)
	if err != nil {
		return fmt.Errorf("Service not found id(%this)", this.Sessionpb.Serverid)
	}
	//cause timeout
	result, err := server.Call("Push", this.Sessionpb.Sessionid, this.Sessionpb.Settings)
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

	if this.Sessionpb.Settings == nil {
		this.Sessionpb.Settings = map[string]string{}
	}
	//this.lock.Lock()
	this.Sessionpb.Settings[key] = value
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
	if this.Sessionpb.Settings == nil {
		this.Sessionpb.Settings = map[string]string{}
	}
	this.Sessionpb.Settings[key] = value
	return this.Push()
}

func (this *session) Get(key string) (result string) {
	if this.Sessionpb.Settings == nil {
		return
	}
	//this.lock.RLock()
	result = this.Sessionpb.Settings[key]
	//this.lock.RUnlock()
	return
}

func (this *session) Remove(key string) (error) {
	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	if this.Sessionpb.Settings == nil {
		this.Sessionpb.Settings=map[string]string{}
	}
	//this.lock.Lock()
	delete(this.Sessionpb.Settings, key)
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
	server, e := this.app.GetServerById(this.Sessionpb.Serverid)
	if e != nil {
		return fmt.Errorf("Service not found id(%this)", this.Sessionpb.Serverid)
	}

	_, err := server.Call("Send", this.Sessionpb.Sessionid, topic, body)
	return err
}

func (this *session) SendBatch(Sessionids string, topic string, body []byte) (int64, error) {
	if this.app == nil {
		return 0, fmt.Errorf("Module.App is nil")
	}
	server, e := this.app.GetServerById(this.Sessionpb.Serverid)
	if e != nil {
		return 0, fmt.Errorf("Service not found id(%s)", this.Sessionpb.Serverid)
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

	server, err := this.app.GetServerById(this.Sessionpb.Serverid)
	if err != nil {
		return fmt.Errorf("Service not found id(%this)", this.Sessionpb.Serverid)
	}

	err = server.CallNR("Send", this.Sessionpb.Sessionid, topic, body)
	return err
}

func (this *session) Close() (error) {
	if this.app == nil {
		return fmt.Errorf("Module.app is nil")
	}

	server, err := this.app.GetServerById(this.Sessionpb.Serverid)
	if err != nil {
		return  fmt.Errorf("Service not found id(%this)", this.Sessionpb.Serverid)
	}

	_, err = server.GetClient().Call("Close", 5, this.Sessionpb.Sessionid)
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
	se := &Sessionpb{
		IP:        this.Sessionpb.IP,
		Network:   this.Sessionpb.Network,
		Userid:    this.Sessionpb.Userid,
		Sessionid: this.Sessionpb.Sessionid,
		Serverid:  this.Sessionpb.Serverid,
		Settings:  this.Sessionpb.Settings,
	}
	//这个要换成本次RPC调用的新Span
	se.Carrier = this.inject()

	s.Sessionpb = se
	return s
}

func (this *session) IsConnect(userId string) (bool, error) {
	if this.app == nil {
		return false, fmt.Errorf("Module.App is nil")
	}
	server, e := this.app.GetServerById(this.Sessionpb.Serverid)
	if e != nil {
		return false, fmt.Errorf("Service not found id(%s)", this.Sessionpb.Serverid)
	}
	results, err := server.Call("IsConnect", this.Sessionpb.Sessionid, userId)
	if err != nil {
		return false, err
	}
	if results != nil && len(results) > 0 {
		if re, ok := results[0].(bool); ok {
			return re, nil
		}
	}
	return false, fmt.Errorf("RPC call [isConnect] get unsupport result")
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
		if this.Sessionpb.Carrier != nil {
			//从已有记录恢复
			clientContext, err := this.extract(this.Sessionpb.Carrier)
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
	this.Sessionpb.Carrier = this.inject()
	return this.span
}

func (this *session) Span() opentracing.Span {
	return this.span
}

func (this *session) TracingCarrier() map[string]string {
	return this.Sessionpb.Carrier
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
	} else if this.app != nil && this.app.GetJudgeGuest() != nil {
		this.judgeGuest = this.app.GetJudgeGuest()
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