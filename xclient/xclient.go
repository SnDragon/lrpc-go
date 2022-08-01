package xclient

import (
	"context"
	"github.com/SnDragon/lrpc-go/client"
	"github.com/SnDragon/lrpc-go/server"
	"io"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opts    []server.OptionFunc
	mu      sync.Mutex
	clients map[string]*client.Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opts ...server.OptionFunc) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opts:    opts,
		clients: make(map[string]*client.Client),
	}
}
func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, c := range xc.clients {
		_ = c.Close()
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) dial(rpcAddr string) (*client.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	c, ok := xc.clients[rpcAddr]
	if ok && !c.IsAvailable() {
		_ = c.Close()
		delete(xc.clients, rpcAddr)
		c = nil
	}
	if c == nil {
		var err error
		c, err = client.XDial(rpcAddr, xc.opts...)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = c
	}
	return c, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceName string, argv, reply interface{}) error {
	c, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return c.Call(ctx, serviceName, argv, reply)
}

func (xc *XClient) Call(ctx context.Context, serviceName string, argv, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceName, argv, reply)
}

func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, argv, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return nil
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	replyDone := reply == nil
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr, ctx, serviceMethod, argv, clonedReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel()
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
