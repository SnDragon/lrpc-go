package codec

func init() {
	RegisterCompressor(CompressTypeNoop, &NoopCompressor{})
}

type NoopCompressor struct {
}

func (n NoopCompressor) Compress(in []byte) (out []byte, err error) {
	return in, nil
}

func (n NoopCompressor) Decompress(in []byte) (out []byte, err error) {
	return in, nil
}
