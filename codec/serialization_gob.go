package codec

import (
	"bytes"
	"encoding/gob"
)

func init() {
	RegisterSerializer(SerializationTypeGob, &GobSerializer{})
}

type GobSerializer struct {
}

func (g GobSerializer) Marshal(in interface{}) ([]byte, error) {
	var buffs = bytes.Buffer{}
	if err := gob.NewEncoder(&buffs).Encode(in); err != nil {
		return nil, err
	}
	return buffs.Bytes(), nil
}

func (g GobSerializer) Unmarshal(data []byte, target interface{}) error {
	var buffs = bytes.NewReader(data)
	return gob.NewDecoder(buffs).Decode(target)
}
