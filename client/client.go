package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SnDragon/lrpc-go/codec"
	"github.com/SnDragon/lrpc-go/server"
	"io"
	"net"
	"sync"
)

type Client struct {
	cc       codec.Codec
	opt      *server.Option
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	pending  map[uint64]*Call
	seq      uint64
	closing  bool
	shutdown bool
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
func (c *Client) Call(serviceName string, argv, reply interface{}) error {
	done := make(chan *Call, 1)
	call := <-c.Go(serviceName, argv, reply, done).Done
	return call.Error
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

func Dial(network, address string, opts ...server.OptionFunc) (client *Client, err error) {
	opt := server.DefaultOption
	for _, optFunc := range opts {
		optFunc(opt)
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
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
