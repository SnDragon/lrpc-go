package codec

import (
	"fmt"
	"testing"
)

func TestGzipCompress_Compress(t *testing.T) {
	compressor := GetCompressor(CompressTypeGzip)
	data := []byte("dfdfdfdfd hello worlddfdfdfdfd hello worlddfdfdfdfd hello worlddfdfdfdfd hello worlddfdfdfdfd hello worldHello 蓝影闪电")
	fmt.Println(len(data))
	out, err := compressor.Compress(data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(out))
	fmt.Println(out)
	o, err := compressor.Decompress(out)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(o))
	fmt.Println(string(o))

}
