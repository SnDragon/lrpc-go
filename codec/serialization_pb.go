package codec

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

func init() {
	RegisterSerializer(SerializationTypePB, &PBSerializer{})
}

type PBSerializer struct {
}

func (P PBSerializer) Marshal(in interface{}) (out []byte, err error) {
	msg, ok := in.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("not protobuf message")
	}
	return proto.Marshal(msg)
}

func (P PBSerializer) Unmarshal(data []byte, target interface{}) error {
	msg, ok := target.(proto.Message)
	if !ok {
		return fmt.Errorf("unmarshal failed,target not protobuf message")
	}
	return proto.Unmarshal(data, msg)
}
