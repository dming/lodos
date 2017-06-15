package baserpc_test

import (
	"sync"
//	"reflect"
//	"time"
	"conf"
	"rpc/base"
//	"fmt"
	"fmt"
	"time"
	"reflect"
)

type me struct {
	name string
	age int
}

func Example() {
	mqInfoC := new(conf.Rabbitmq)
	mqInfoC.Uri = "amqp://guest:guest@127.0.0.1"
	mqInfoC.Exchange = "testEx"
	mqInfoC.ExchangeType = "direct"
	mqInfoC.BindingKey = "testKey"
	mqInfoC.ConsumerTag = "testTag"
	mqInfoC.Queue = "testQ"

	mqInfoS := new(conf.Rabbitmq)
	mqInfoS.Uri = "amqp://guest:guest@127.0.0.1"
	mqInfoS.Exchange = "testEx"
	mqInfoS.ExchangeType = "direct"
	mqInfoS.BindingKey = "testKey"
	mqInfoS.ConsumerTag = "testTag"
	mqInfoS.Queue = "testQ"

	s := baserpc.NewServer()
	mqServer, err := baserpc.NewMqServer(nil, mqInfoS, s.ChanCall)
	if err != nil {
		fmt.Println("new server error")
		//return
	}
	s.AttachMqServer(mqServer)

	var wg sync.WaitGroup
	wg.Add(1)

	// goroutine 1
	// the function should be as bellow :
	// func funcName (inputs) (outputs, error)
	// PS: error in output is require
	go func() {
		s.Register("f0", func() (error) {
			//fmt.Println("Call f0")
			return nil
		})

		s.Register("f1", func() (interface{}, error) {
			//fmt.Println("Call f1")
			time.Sleep(time.Second * 0)
			return 1, nil
		})

		s.RegisterGo("fn", func() ([]interface{}, error) {
			//fmt.Println("Call fn")
			r := make([]interface{}, 3)
			r[0] = "hello"
			ll := new(me)
			ll.name = "dming"
			ll.age = 18
			r[1]= ll
			r[2] = 3
			return r, nil
		})
		//first int, second int
		s.Register("add", func(args ...interface{}) (interface{}, error) {
			//fmt.Printf("%v + %v = %v\n", args[0].(int), args[1].(int), args[0].(int) + args[1].(int))
			//return first + second
			//fmt.Println("Call add")
			if (len(args) < 2) {
				return nil, fmt.Errorf("need 2 args")
			} else {
				for i := 0; i < 2; i++ {
					if reflect.TypeOf(args[i]).Kind() != reflect.Int {
						return nil, fmt.Errorf(" args need to be int")
					}
				}
			}

			return args[0].(int) + args[1].(int), nil
		})

		wg.Done()
	}()

	wg.Wait()
	wg.Add(1)

	// goroutine 2
	go func() {
		// the args of '10' is the cache for async chanel , c.AsynCall count should not larger than it.
		c := baserpc.NewClient()
		c.AttachChanCall(s.ChanCall)
		mqClient, err := baserpc.NewMqClient(nil, mqInfoC)
		if err != nil {
			fmt.Println("create mq client fail !!")
		} else {
			c.AttachMqClient(mqClient)
		}
		// sync
		r0, err := c.Call("f0", 1)
		if err != nil || r0.Err != nil {
			fmt.Println(err)
			fmt.Println(r0.Err)
		}

		r1, err := c.Call("f1", 1)
		if err != nil || r1.Err != nil {
			fmt.Printf("err of r1 is %s\n", err)
			fmt.Printf("err of r1.err is %s\n", r1.Err.Error())
		} else {
			fmt.Println(r1.Ret[0])
		}

		rn, err := c.Call("fn", 1)
		if err != nil || rn.Err != nil {
			fmt.Println(err)
			fmt.Println(rn.Err)
		} else {
			rn := rn.Ret[0].([]interface{})
			fmt.Println(rn[0], rn[1], rn[2])
			//fmt.Println(rn)
		}

		ra, err := c.Call("add",1, 1, 2)
		if err != nil || ra.Err != nil {
			fmt.Println(err)
			fmt.Println(ra.Err)
		} else {
			fmt.Println(ra.Ret[0])
		}

		// asyn
		//chanF0, _ := c.AsynCall("f0")

		//chanF1, _ := c.AsynCall("f1")

		//chanFn, _ := c.AsynCall("fn")

		ak := make([]interface{}, 2)
		ak[0] = 1
		ak[1] = 2
		//chanAdd, _ := c.AsynCall("add", 1, 2)

/*
		// client Asyn Call的直接返回chan让client自行处理
		<- chanF0
		close(chanF0)

		ri1 := <-chanF1
		fmt.Println(ri1.Ret[0])
		close(chanF1)

		rin := <-chanFn
		arin := rin.Ret[0].([]interface{})
		fmt.Println(arin[0], arin[1], arin[2])
		close(chanFn)

		riAdd := <-chanAdd
		fmt.Println(riAdd.Ret[0])
		close(chanAdd)
*/
		wg.Done()
	}()




	wg.Wait()

	// Output:
	// 1
	// hello &{dming 18} 3
	// 3
	// 1
	// hello &{dming 18} 3
	// 3

}
