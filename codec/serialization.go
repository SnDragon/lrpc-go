package codec

import (
	"fmt"
)

type Serializer interface {
	Marshal(in interface{}) (out []byte, err error)
	Unmarshal(data []byte, target interface{}) error
}

const (
	SerializationTypePB   = 0
	SerializationTypeJson = 1
	SerializationTypeXml  = 2
	SerializationTypeGob  = 3
)

var serializers = make(map[int]Serializer)

func RegisterSerializer(serializationType int, serializer Serializer) {
	serializers[serializationType] = serializer
}

func GetSerializer(serializationType int) Serializer {
	return serializers[serializationType]
}

func Marshal(serializationType int, in interface{}) (out []byte, err error) {
	if in == nil {
		return nil, nil
	}
	serializer := GetSerializer(serializationType)
	if serializer == nil {
		return nil, fmt.Errorf("unsupported serializationType: %v", serializationType)
	}
	return serializer.Marshal(in)
}

func Unmarshal(serializationType int, data []byte, target interface{}) error {
	if len(data) == 0 {
		return nil
	}
	if target == nil {
		return nil
	}
	serializer := GetSerializer(serializationType)
	if serializer == nil {
		return fmt.Errorf("unsupported serializationType: %v", serializationType)
	}
	return serializer.Unmarshal(data, target)
}
