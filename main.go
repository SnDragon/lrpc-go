package main

import (
	"context"
	"fmt"
	"github.com/SnDragon/lrpc-go/client"
	"github.com/SnDragon/lrpc-go/server"
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
	if err := s.Accept(lis); err != nil {
		panic(err)
	}
}

func main() {
	addr := make(chan string)
	go startServer(addr)
	c, err := client.Dial("tcp", <-addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = c.Close()
	}()
	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			if err := c.Call(ctx, "Foo.Sum", args, &reply); err != nil {
				fmt.Println("Call err:", err)
				return
			}
			//fmt.Println("call reply:", reply)
			fmt.Printf("%d + %d = %d\n", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}
