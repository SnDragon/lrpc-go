package codec

import "encoding/json"

func init() {
	RegisterSerializer(SerializationTypeJson, &JsonSerializer{})
}

type JsonSerializer struct {
}

func (s *JsonSerializer) Unmarshal(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

func (s *JsonSerializer) Marshal(in interface{}) (out []byte, err error) {
	return json.Marshal(in)
}
