package gate

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/dming/lodos/module"
	log "github.com/dming/lodos/mlog"
)

type session struct {
	app       module.AppInterface
	sessionpb   *sessionpb
}

func NewSession(app module.AppInterface, data []byte) (Session,error) {
	agent:=&session{
		app:app,
	}
	se := &sessionpb{}
	err := proto.Unmarshal(data, se)
	if err != nil {
		return nil,err
	}    // 测试结果
	agent.sessionpb = se
	return agent,nil
}

func NewSessionByMap(app module.AppInterface, data map[string]interface{}) (Session,error) {
	agent:=&session{
		app:app,
		sessionpb:new(sessionpb),
	}
	err := agent.updateMap(data)
	if err != nil{
		return nil, err
	}
	return agent, nil
}

func (session *session) GetIP() string {
	return session.sessionpb.GetIP()
}

func (session *session) GetNetwork() string {
	return session.sessionpb.GetNetwork()
}

func (session *session) GetUserid() string {
	return session.sessionpb.GetUserid()
}

func (session *session) GetSessionid() string {
	return session.sessionpb.GetSessionid()
}

func (session *session) GetServerid() string {
	return session.sessionpb.GetServerid()
}

func (session *session) GetSettings() map[string]string {
	return session.sessionpb.GetSettings()
}


func (session *session)SetIP(ip string){
	session.sessionpb.IP=ip
}
func (session *session)SetNetwork(network string){
	session.sessionpb.Network=network
}
func (session *session)SetUserid(userid string){
	session.sessionpb.Userid=userid
}
func (session *session)SetSessionid(sessionid string){
	session.sessionpb.Sessionid=sessionid
}
func (session *session)SetServerid(serverid string){
	session.sessionpb.Serverid=serverid
}
func (session *session)SetSettings(settings map[string]string){
	session.sessionpb.Settings=settings
}

func (session *session) updateMap(s map[string]interface{}) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("Error on update map, [%s]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	Userid := s["Userid"]
	if Userid != nil {
		session.sessionpb.Userid = Userid.(string)
	}
	IP := s["IP"]
	if IP != nil {
		session.sessionpb.IP = IP.(string)
	}
	Network := s["Network"]
	if Network != nil {
		session.sessionpb.Network = Network.(string)
	}
	Sessionid := s["Sessionid"]
	if Sessionid != nil {
		session.sessionpb.Sessionid = Sessionid.(string)
	}
	Serverid := s["Serverid"]
	if Serverid != nil {
		session.sessionpb.Serverid = Serverid.(string)
	}
	Settings := s["Settings"]
	if Settings != nil {
		session.sessionpb.Settings = Settings.(map[string]string)
	}
	return err
}

func (session *session) update (s Session) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("Error on update, [%s]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	session.sessionpb.Userid = s.GetUserid()

	session.sessionpb.IP = s.GetIP()

	session.sessionpb.Network = s.GetNetwork()

	session.sessionpb.Sessionid = s.GetSessionid()

	session.sessionpb.Serverid = s.GetServerid()

	session.sessionpb.Settings = s.GetSettings()
	return err
}

func (session *session) Serializable() ([]byte, error){
	data, err := proto.Marshal(session.sessionpb)
	if err != nil {
		return nil, err
	}    // 进行解码
	return data, nil
}

// call Update in handler
func (session *session) Update() (error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("error in update [%s]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}

	server, err := session.app.GetServerById(session.sessionpb.GetServerid())
	if err != nil {
		return fmt.Errorf("Service not found id(%s)", session.sessionpb.GetServerid())
	}

	result, err := server.Call("Update", 5, session.sessionpb.GetSessionid())
	if err != nil {
		return err
	}
	if result != nil && len(result.Ret) > 0 {
		//绑定成功,重新更新当前Session
		session.update(result.Ret[0].(Session))
	}
	return err
}

func (session *session) Bind(Userid string) (error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("error in Bind [%s]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}

	server, err := session.app.GetServerById(session.sessionpb.Serverid)
	if err != nil {
		return fmt.Errorf("Service not found id(%s)", session.sessionpb.Serverid)
	}

	result, err := server.Call("Bind", 5,  session.sessionpb.Sessionid, Userid)
	log.Debug("in Bind, result is : %v", result)
	if err != nil {
		return err
	}
	if result != nil && len(result.Ret) > 0 {
		session.update(result.Ret[0].(Session))
	}
	return err
}

func (session *session) UnBind() (error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Error("error in Bind [%s]", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}
	server, err := session.app.GetServerById(session.sessionpb.Serverid)
	if err != nil {
		return fmt.Errorf("Service not found id(%s), err is %s", session.sessionpb.Serverid, err)
	}

	result, err := server.GetClient().Call("UnBind", 5, session.sessionpb.Sessionid)
	if err != nil {
		return err
	}
	if result != nil  && len(result.Ret) > 0 {
		//绑定成功,重新更新当前Session
		session.update(result.Ret[0].(Session))
	}
	return err
}

func (session *session) Push() (error) {
	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}

	server, e := session.app.GetServerById(session.sessionpb.Serverid)
	if e != nil {
		return fmt.Errorf("Service not found id(%s)", session.sessionpb.Serverid)
	}

	result, err := server.GetClient().Call("Push", 5, session.sessionpb.Sessionid, session.sessionpb.Settings)
	if err != nil {
		return err
	}
	if result != nil  && len(result.Ret) > 0 {
		//绑定成功,重新更新当前Session
		session.update(result.Ret[0].(Session))
	}
	return err
}

func (session *session) Set(key string, value string) (error) {
	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}

	if session.sessionpb.Settings == nil {
		session.sessionpb.Settings=map[string]string{}
	}
	session.sessionpb.Settings[key] = value
	//server,e:=session.app.GetServersById(session.Serverid)
	//if e!=nil{
	//	err=fmt.Sprintf("Service not found id(%s)",session.Serverid)
	//	return
	//}
	//result,err:=server.Call("Set",session.Sessionid,key,value)
	//if err==""{
	//	if result!=nil{
	//		//绑定成功,重新更新当前Session
	//		session.update(result.(map[string]interface {}))
	//	}
	//}
	return nil
}

func (session *session) Get(key string) (result string) {
	if session.sessionpb.Settings == nil {
		return
	}
	result = session.sessionpb.Settings[key]
	return
}

func (session *session) Remove(key string) (error) {
	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}

	if session.sessionpb.Settings == nil {
		session.sessionpb.Settings=map[string]string{}
	}
	delete(session.sessionpb.Settings, key)
	//server,e:=session.app.GetServersById(session.Serverid)
	//if e!=nil{
	//	err=fmt.Sprintf("Service not found id(%s)",session.Serverid)
	//	return
	//}
	//result,err:=server.Call("Remove",session.Sessionid,key)
	//if err==""{
	//	if result!=nil{
	//		//绑定成功,重新更新当前Session
	//		session.update(result.(map[string]interface {}))
	//	}
	//}
	return nil
}
func (session *session) Send(topic string, body []byte) (error) {
	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}
	server, e := session.app.GetServerById(session.sessionpb.Serverid)
	if e != nil {
		return fmt.Errorf("Service not found id(%s)", session.sessionpb.Serverid)
	}
	_, err := server.Call("Send", 5, session.sessionpb.Sessionid, topic, body)
	return err
}

func (session *session) SendNR(topic string, body []byte) (error) {
	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}

	server, err := session.app.GetServerById(session.sessionpb.Serverid)
	if err != nil {
		return fmt.Errorf("Service not found id(%s)", session.sessionpb.Serverid)
	}

	_, err = server.GetClient().Call("Send", 5, session.sessionpb.Sessionid, topic, body)
	if err != nil {
		return err
	}
	return nil
}

func (session *session) Close() (error) {
	if session.app == nil {
		return fmt.Errorf("Module.App is nil")
	}

	server, err := session.app.GetServerById(session.sessionpb.Serverid)
	if err != nil {
		return  fmt.Errorf("Service not found id(%s)", session.sessionpb.Serverid)
	}

	_, err = server.GetClient().Call("Close", 5, session.sessionpb.Sessionid)
	return err
}
