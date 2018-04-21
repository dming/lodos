package base

import "github.com/dming/seqsvr/seq"

// allocsvr流程：
// 1. 从存储系统里取出路由表.(从本地set里寻找到一个可用的storesvr)
// 2. 通过本机Addr信息到路由表里获取当前服务号段列表
// 3. 启动租约服务
// 4. 从存储系统里取出max_seqs集合 ->4
// 5. 开始对客户端提供服务

//4
/*
 1. 内存中储存最近一个分配出去的sequence：cur_seq，以及分配上限：max_seq
 2. 分配sequence时，将cur_seq++，同时与分配上限max_seq比较：
    如果cur_seq > max_seq，将分配上限提升一个步长max_seq += step，并持久化max_seq
 3. 重启时，读出持久化的max_seq，赋值给cur_seq
 */

type allocSvr struct{
	settings map[string]string //init settings while on init
	router seq.RouterTable //get table from storesvr
	store seq.StoreSvr // or a connection to storesvr
	managers []seq.SectionManager
	leaseClerk seq.LeaseClerk

	//mqtt MQTTClient
	done_chan chan bool
}

func (this *allocSvr) Run() {
	panic("implement me")
}

func (this *allocSvr) OnDestroy() {
	this.done_chan <- true
}

func (this *allocSvr) OnInit(settings map[string]string) {
	this.settings = settings
	this.store = NewStoreSvr(settings["storeList"])//create a [connection to] store
	this.router = this.store.GetRouterTable() // get router table from store svr

	//todo: do something to lease clerk
	//code
	this.leaseClerk = this

	this.initSections(this.router) //create sections depending on self host [in settings]
	//start service
	go this.listen_loop(this.done_chan)
	return
}

func (this *allocSvr) GetRouter() {
	panic("implement me")
}

func (this *allocSvr) GetSeq(id int) (seq seq.Sequence, err error) {
	panic("implement me")
}

func (this *allocSvr) initSections(router seq.RouterTable) {

}

// mqtt listener, get request from client
func (this *allocSvr) listen_loop(done_chan chan bool) {
	panic("implement me")
	for {
		switch {
		case <-done_chan:
			return
		//case ://get request from clients
		}
	}
}

//implement of lease clerk
func (this *allocSvr) OnLeaseValid() {
	panic("implement me")
}

func (this *allocSvr) OnLeaseUpdated() {
	panic("implement me")
}

func (this *allocSvr) OnLeaseInvalid() {
	panic("implement me")
}

func (this *allocSvr) Start() {
	panic("implement me")
}

func (this *allocSvr) Stop() {
	panic("implement me")
}