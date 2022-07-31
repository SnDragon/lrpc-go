package main

import (
	"fmt"
	"github.com/SnDragon/lrpc-go/client"
	"github.com/SnDragon/lrpc-go/server"
	"net"
	"sync"
	"time"
)

func startServer(addr chan string) {
	s := server.NewServer()
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
			argv := fmt.Sprintf("rpc req:%d", i)
			var reply string
			if err := c.Call("RPC.Hello", argv, &reply); err != nil {
				fmt.Println("Call err:", err)
				return
			}
			fmt.Println("call reply:", reply)
		}(i)
	}
	wg.Wait()
}
