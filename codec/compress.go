package codec

import "fmt"

type Compressor interface {
	Compress(in []byte) (out []byte, err error)
	Decompress(in []byte) (out []byte, err error)
}

const (
	CompressTypeNoop   = 0
	CompressTypeGzip   = 1
	CompressTypeSnappy = 2
	CompressTypeZlib   = 3
)

var compressors = make(map[int]Compressor)

func RegisterCompressor(compressType int, s Compressor) {
	compressors[compressType] = s
}

func GetCompressor(compressType int) Compressor {
	return compressors[compressType]
}

func Compress(compressType int, in []byte) (out []byte, err error) {
	if len(in) == 0 {
		return nil, nil
	}
	compressor := GetCompressor(compressType)
	if compressor == nil {
		return nil, fmt.Errorf("compressor not registered")
	}
	return compressor.Compress(in)
}

func Decompress(compressType int, in []byte) (out []byte, err error) {
	if len(in) == 0 {
		return nil, nil
	}
	compressor := GetCompressor(compressType)
	if compressor != nil {
		return nil, fmt.Errorf("compressor not registered")
	}
	return compressor.Decompress(in)
}
