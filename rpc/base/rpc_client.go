package baserpc

import (
	"strconv"
	"time"
	"fmt"
	"github.com/dming/lodos/rpc"
	log "github.com/dming/lodos/mlog"
)

// client should only know about the channel connected to server, \
// but dont know who is the server
type client struct {
	ChanCall        chan *rpc.CallInfo // send info
	chanSyncRet     chan *rpc.RetInfo  // get info
	mqClient rpc.MqClient
	//wg sync.WaitGroup
	flag int64
}

func NewClient() rpc.Client {
	c := new(client)
	c.chanSyncRet = make(chan *rpc.RetInfo, 1)
	//c.ChanCall = chanCall
	c.flag = 0
	//c.mqClient, err = NewMqClient(info)
	return c
}

func (c *client) AttachMqClient(mqClient rpc.MqClient) {
	c.mqClient = mqClient
}

func (c *client) AttachChanCall(chanCall chan *rpc.CallInfo) {
	c.ChanCall = chanCall
}

func (c *client) Call(id string, timeout int, args ...interface{}) (*rpc.RetInfo , error) {
	ci := new(rpc.CallInfo)
	ci.Id = id
	ci.Args = args
	if timeout <= 0 || timeout > 120 {
		timeout = 10
	}
	ticker := time.NewTicker(time.Second * 10)

	if c.ChanCall != nil {
		chanSyncRet := make(chan *rpc.RetInfo, 1)
		ci.ChanRet = chanSyncRet
		ci.ReplyTo = ""
		err := c.call(ci)
		if err != nil {
			return nil, err
		}

		select {
		case ri := <-chanSyncRet:
			close(chanSyncRet)
			return ri, ri.Err
		case <-ticker.C:
			ticker.Stop()
			ri := &rpc.RetInfo{Err: fmt.Errorf("function %s timeout!!!", ci.Id)}
			return ri, ri.Err
		}
	} else if c.mqClient != nil {
		chanSyncRet := make(chan *rpc.RetInfo, 1)
		ci.ChanRet = chanSyncRet
		//ci.ReplyTo = c.mqClient.Consumer.callback_queue
		c.flag++
		ci.Flag = strconv.FormatInt(c.flag, 10)
		err := c.mqClient.Call(ci)
		if err != nil {
			return nil, err
		}

		select {
		case ri := <-chanSyncRet:
			close(chanSyncRet)
			return ri, ri.Err
		case <-ticker.C:
			ticker.Stop()
			ri := &rpc.RetInfo{Err: fmt.Errorf("function %s timeout!!!", ci.Id)}
			return ri, ri.Err
		}
	} else {
		return nil, fmt.Errorf("not local client and mq client")
	}
}

func (c *client) call(ci *rpc.CallInfo) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			log.Error(err.Error())
		}
	}()
	
	c.ChanCall <- ci

	return
}


func (c *client) AsynCall(id string, args...interface{}) (chan *rpc.RetInfo, error) {

	ci := new(rpc.CallInfo)
	ci.Id = id
	ci.Args = args

	//get return info by chan chanAsynRet
	if c.ChanCall != nil {
		chanAsynRet := make(chan *rpc.RetInfo, 1)
		ci.ChanRet = chanAsynRet
		ci.ReplyTo = ""
		err := c.asynCall(ci)
		return chanAsynRet, err
	} else if c.mqClient != nil {
		chanAsynRet := make(chan *rpc.RetInfo, 1)
		ci.ChanRet = chanAsynRet
		//ci.ReplyTo = c.mqClient.Consumer.callback_queue
		c.flag++
		ci.Flag = strconv.FormatInt(c.flag, 10)
		err := c.asynCall(ci)
		return chanAsynRet, err
	} else {
		return nil, fmt.Errorf("not local client and mq client")
	}
}

func (c *client) asynCall(ci *rpc.CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
		return
	}()
	//get return info by chan chanAsynRet
	return c.call(ci)
}



func (c *client) Close() error {
	if c.chanSyncRet != nil {
		close(c.chanSyncRet)
	}

	if c.mqClient != nil {
		return c.mqClient.Done()
	}
	return nil
}
