package main

import (
	"encoding/json"
	"fmt"
	"github.com/SnDragon/lrpc-go/codec"
	"github.com/SnDragon/lrpc-go/server"
	"net"
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
	conn, err := net.Dial("tcp", <-addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = conn.Close()
	}()
	time.Sleep(time.Second)
	_ = json.NewEncoder(conn).Encode(server.DefaultOption)
	gobCodec := codec.NewCodecTypeGob(conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "RPC.Hello",
			Seq:           i,
		}
		_ = gobCodec.Write(h, "Hello from longerwu")
		_ = gobCodec.ReadHeader(h)
		var reply string
		_ = gobCodec.ReadBody(&reply)
		fmt.Println("reply:", reply)
	}
}
