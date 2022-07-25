package codec

import "io"

type Header struct {
	ServiceMethod string // `service.method`
	Seq           int
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(h *Header) error
	ReadBody(body interface{}) error
	Write(h *Header, body interface{}) error
}

type CodecType int

const (
	CodecTypeGob  CodecType = 1
	CodecTypeJson CodecType = 2 // unimplemented
)

var CodecTypeMap = make(map[CodecType]NewCodecType)

type NewCodecType func(closer io.ReadWriteCloser) Codec

func init() {
	CodecTypeMap[CodecTypeGob] = NewCodecTypeGob
}
