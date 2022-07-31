package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SnDragon/lrpc-go/codec"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int             `json:"magic_number"`
	CodecType   codec.CodecType `json:"codec_type"`
}

type OptionFunc func(option *Option)

func WithCodecType(codecType codec.CodecType) OptionFunc {
	return func(option *Option) {
		option.CodecType = codecType
	}
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.CodecTypeGob,
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Register(rcvr interface{}) error {
	service := newService(rcvr)
	if _, existed := s.serviceMap.LoadOrStore(service.name, service); existed {
		return errors.New("rpc: service already define:" + service.name)
	}
	return nil
}

func (s *Server) findService(serviceMethod string) (svr *service, m *methodType, err error) {
	idx := strings.Index(serviceMethod, ".")
	if idx <= 0 {
		err = fmt.Errorf("server err: serviceMethod %s not found", serviceMethod)
		return
	}
	svrName, methodName := serviceMethod[0:idx], serviceMethod[idx+1:]
	svrInter, ok := s.serviceMap.Load(svrName)
	if !ok {
		err = fmt.Errorf("server err: service %s not found", svrName)
		return
	}
	svr = svrInter.(*service)
	m = svr.methods[methodName]
	if m == nil {
		err = fmt.Errorf("server err: method %s not found", methodName)
		return
	}
	return
}

func (s *Server) Accept(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		fmt.Println("ServeConn err:", err)
		return
	}
	// 校验option magicNumber,codec等
	if opt.MagicNumber != MagicNumber {
		fmt.Println("err:", fmt.Errorf("invalid magicNumber: %v", opt.MagicNumber))
		return
	}
	f := codec.CodecTypeMap[opt.CodecType]
	if f == nil {
		fmt.Println("err:", fmt.Errorf("invalid codecType: %v", opt.CodecType))
		return
	}
	s.serveCodec(f(conn))
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

func (s *Server) serveCodec(c codec.Codec) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for {
		req, err := s.readRequest(c)
		if err != nil {
			// EOF
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				fmt.Println("read done...")
				break
			}
			fmt.Println("readRequest err:", err)
			// 发送错误消息
			req.h.Error = err.Error()
			s.sendResponse(c, req.h, invalidRequest, mu)
			return
		}
		wg.Add(1)
		go s.handleRequest(c, req, wg, mu)
	}
	wg.Wait()
}

type Request struct {
	h            *codec.Header
	argv, replyv reflect.Value
	svr          *service
	mType        *methodType
}

func (s *Server) readRequest(c codec.Codec) (r *Request, err error) {
	h := &codec.Header{}
	if err := c.ReadHeader(h); err != nil {
		fmt.Println("ReadHeader err:", err)
		return nil, err
	}
	r = &Request{
		h: h,
	}
	r.svr, r.mType, err = s.findService(h.ServiceMethod)
	r.argv = r.mType.newArgv()
	r.replyv = r.mType.newReplyv()
	argvi := r.argv.Interface()
	if r.argv.Type().Kind() != reflect.Pointer {
		argvi = r.argv.Addr().Interface()
	}
	if err := c.ReadBody(argvi); err != nil {
		fmt.Println("ReadBody err:", err)
		return nil, err
	}
	return r, nil
}

func (s *Server) handleRequest(c codec.Codec, req *Request, wg *sync.WaitGroup, mu *sync.Mutex) error {
	defer func() {
		wg.Done()
	}()
	fmt.Println(req.h, req.argv.Elem())
	//req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	// 通过反射调用对应服务等逻辑处理方法
	if err := req.svr.call(req.mType, req.argv, req.replyv); err != nil {
		return err
	}
	return s.sendResponse(c, req.h, req.replyv.Interface(), mu)
}

func (s *Server) sendResponse(c codec.Codec, h *codec.Header, body interface{}, mu *sync.Mutex) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.Write(h, body); err != nil {
		fmt.Println("sendResponse err:", err)
		return err
	}
	return nil
}
