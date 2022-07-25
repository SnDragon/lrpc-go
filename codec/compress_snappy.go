package codec

import (
	"bytes"
	"github.com/golang/snappy"
	"io/ioutil"
)

func init() {
	RegisterCompressor(CompressTypeSnappy, &SnappyCompressor{})
}

type SnappyCompressor struct {
}

func (s SnappyCompressor) Compress(in []byte) ([]byte, error) {
	var b bytes.Buffer
	w := snappy.NewBufferedWriter(&b)
	if _, err := w.Write(in); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (s SnappyCompressor) Decompress(in []byte) (out []byte, err error) {
	r := snappy.NewReader(bytes.NewReader(in))
	return ioutil.ReadAll(r)
}
