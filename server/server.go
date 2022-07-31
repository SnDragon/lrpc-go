package server

import (
	"encoding/json"
	"fmt"
	"github.com/SnDragon/lrpc-go/codec"
	"io"
	"net"
	"reflect"
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
}

func NewServer() *Server {
	return &Server{}
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
	h           *codec.Header
	argv, reply reflect.Value
}

func (s *Server) readRequest(c codec.Codec) (*Request, error) {
	h := &codec.Header{}
	if err := c.ReadHeader(h); err != nil {
		fmt.Println("ReadHeader err:", err)
		return nil, err
	}
	r := &Request{
		h: h,
	}
	// TODO
	r.argv = reflect.New(reflect.TypeOf(""))
	if err := c.ReadBody(r.argv.Interface()); err != nil {
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
	req.reply = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	return s.sendResponse(c, req.h, req.reply.Interface(), mu)
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
