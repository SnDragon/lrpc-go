package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SnDragon/lrpc-go/codec"
	"github.com/SnDragon/lrpc-go/server"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	cc       codec.Codec
	opt      *server.Option
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	pending  map[uint64]*Call
	seq      uint64
	closing  bool // 客户端主动关闭
	shutdown bool // 出现错误被动关闭
}

var _ io.Closer = (*Client)(nil)

var ErrShutDown = errors.New("connection is shutdown")

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return ErrShutDown
	}
	c.closing = true
	return c.cc.Close()
}

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.closing && !c.shutdown
}

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutDown
	}
	call.Seq = c.seq
	c.pending[c.seq] = call
	c.seq++
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	c.mu.Unlock()
	target := c.pending[seq]
	delete(c.pending, seq)
	return target
}

func (c *Client) terminalCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
	c.shutdown = true
}

func (c *Client) send(call *Call) {
	c.sending.Lock()
	defer c.sending.Unlock()
	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}
	c.header.Seq = seq
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Error = ""
	if err := c.cc.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (c *Client) Go(serviceName string, argv, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceName,
		Args:          argv,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)
	return call
}
func (c *Client) Call(ctx context.Context, serviceName string, argv, reply interface{}) error {
	done := make(chan *Call, 1)
	call := c.Go(serviceName, argv, reply, done)
	select {
	case <-ctx.Done():
		c.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}

func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = c.cc.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			err = c.cc.ReadBody(nil)
			call.Error = errors.New(h.Error)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("error reading body:" + err.Error())
			}
			call.done()
		}
	}
	c.terminalCalls(err)
}

type clientResult struct {
	client *Client
	err    error
}

type newClientFunc func(conn net.Conn, opt *server.Option) (*Client, error)

func DialTimeout(f newClientFunc, network, address string, opts ...server.OptionFunc) (client *Client, err error) {
	opt := server.DefaultOption
	for _, optFunc := range opts {
		optFunc(&opt)
	}
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	ch := make(chan *clientResult)
	go func() {
		c, err := f(conn, &opt)
		ch <- &clientResult{
			client: c,
			err:    err,
		}
	}()
	if opt.ConnectTimeout == 0 {
		ret := <-ch
		return ret.client, ret.err
	}
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case ret := <-ch:
		return ret.client, ret.err
	}
}

func Dial(network, address string, opts ...server.OptionFunc) (client *Client, err error) {
	return DialTimeout(NewClient, network, address, opts...)
}

func NewClient(conn net.Conn, opt *server.Option) (*Client, error) {
	f := codec.CodecTypeMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codecType:%v", opt.CodecType)
		fmt.Println("NewClient err:", err)
		return nil, err
	}
	// send option
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		fmt.Println("send opt err:", err)
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(codec codec.Codec, opt *server.Option) *Client {
	c := &Client{
		seq:     1,
		cc:      codec,
		opt:     opt,
		pending: map[uint64]*Call{},
	}
	go c.receive()
	return c
}

func NewHTTPClient(conn net.Conn, opt *server.Option) (*Client, error) {
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", server.DefaultRPCPath))
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == server.Connected {
		return NewClient(conn, opt)
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	return nil, err
}

func DialHTTP(network, address string, opts ...server.OptionFunc) (client *Client, err error) {
	return DialTimeout(NewHTTPClient, network, address, opts...)
}

// XDial calls different functions to connect to a RPC server
// according the first parameter rpcAddr.
// rpcAddr is a general format (protocol@addr) to represent a rpc server
// eg, http@10.0.0.1:7001, tcp@10.0.0.1:9999, unix@/tmp/geerpc.sock
func XDial(rpcAddr string, opts ...server.OptionFunc) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format '%s', expect protocol@addr", rpcAddr)
	}
	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		return Dial(protocol, addr, opts...)
	}
}
