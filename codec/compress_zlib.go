package codec

import (
	"bytes"
	"compress/zlib"
	"io"
)

func init() {
	RegisterCompressor(CompressTypeZlib, &ZlibCompressor{})
}

type ZlibCompressor struct {
}

func (z ZlibCompressor) Compress(in []byte) (out []byte, err error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	if _, err := w.Write(in); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (z ZlibCompressor) Decompress(in []byte) ([]byte, error) {
	b := bytes.NewReader(in)
	var out bytes.Buffer
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
