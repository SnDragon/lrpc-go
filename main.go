package main

import (
	"context"
	"fmt"
	"github.com/SnDragon/lrpc-go/server"
	"github.com/SnDragon/lrpc-go/xclient"
	"log"
	"net"
	"sync"
	"time"
)

type Foo int
type Args struct {
	Num1, Num2 int
}

func (f *Foo) Sum(args *Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f *Foo) Sleep(args *Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addr chan string) {
	s := server.NewServer()
	var foo Foo
	if err := s.Register(&foo); err != nil {
		panic("register err:" + err.Error())
	}
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	addr <- lis.Addr().String()
	fmt.Println("server started at:", lis.Addr().String())
	//s.HandleHTTP()
	//_ = http.Serve(lis, nil)
	if err := s.Accept(lis); err != nil {
		panic(err)
	}
}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// expect 2 - 5 timeout
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {
	//addr := make(chan string)
	//go startServer(addr)
	//c, err := client.Dial("tcp", <-addr)
	//c, err := client.DialHTTP("tcp", <-addr)
	//if err != nil {
	//	panic(err)
	//}
	//defer func() {
	//	_ = c.Close()
	//}()
	//time.Sleep(time.Second)
	//var wg sync.WaitGroup
	//for i := 0; i < 5; i++ {
	//	wg.Add(1)
	//	go func(i int) {
	//		defer wg.Done()
	//		args := &Args{Num1: i, Num2: i * i}
	//		var reply int
	//		ctx, _ := context.WithTimeout(context.Background(), time.Second)
	//		if err := c.Call(ctx, "Foo.Sum", args, &reply); err != nil {
	//			fmt.Println("Call err:", err)
	//			return
	//		}
	//		//fmt.Println("call reply:", reply)
	//		fmt.Printf("%d + %d = %d\n", args.Num1, args.Num2, reply)
	//	}(i)
	//}
	//wg.Wait()
	//time.Sleep(time.Minute * 10)
	log.SetFlags(0)
	ch1 := make(chan string)
	ch2 := make(chan string)
	// start two servers
	go startServer(ch1)
	go startServer(ch2)

	addr1 := <-ch1
	addr2 := <-ch2

	time.Sleep(time.Second)
	call(addr1, addr2)
	broadcast(addr1, addr2)
}
