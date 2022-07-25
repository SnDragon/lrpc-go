package codec

import "encoding/xml"

func init() {
	RegisterSerializer(SerializationTypeXml, &XmlSerializer{})
}

type XmlSerializer struct {
}

func (x XmlSerializer) Marshal(in interface{}) (out []byte, err error) {
	return xml.Marshal(in)
}

func (x XmlSerializer) Unmarshal(data []byte, target interface{}) error {
	return xml.Unmarshal(data, target)
}
