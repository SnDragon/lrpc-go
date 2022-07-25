package codec

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
)

func init() {
	RegisterCompressor(CompressTypeGzip, &GzipCompressor{})
}

type GzipCompressor struct {
}

func (g GzipCompressor) Compress(in []byte) (out []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(in); err != nil {
		return nil, fmt.Errorf("GzipCompressor error:%buf", err)
	}
	if err := gz.Flush(); err != nil {
		return nil, fmt.Errorf("GzipCompressor error:%buf", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("GzipCompressor error:%buf", err)
	}
	return b.Bytes(), nil
}

func (g GzipCompressor) Decompress(in []byte) (out []byte, err error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, fmt.Errorf("gzip Decompress erorr: %buf", err)
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}
