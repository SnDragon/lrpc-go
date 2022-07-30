package codec

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	dec  *gob.Decoder
	enc  *gob.Encoder
	buf  *bufio.Writer
}

func NewCodecTypeGob(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
		buf:  buf,
	}
}

func (c GobCodec) Close() error {
	return c.conn.Close()
}

func (c GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c GobCodec) ReadBody(b interface{}) error {
	return c.dec.Decode(b)
}

func (c GobCodec) Write(h *Header, body interface{}) error {
	defer func() {
		if err := c.buf.Flush(); err != nil {
			_ = c.Close()
		}
	}()
	if err := c.enc.Encode(h); err != nil {
		fmt.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		fmt.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

// 确保 struct GobCodec 实现了接口 Codec。
// 这样 IDE 和编译期间就可以检查，而不是等到使用的时候。
var _ Codec = (*GobCodec)(nil)

// or var _ Codec = &GobCodec{}
