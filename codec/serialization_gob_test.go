package codec

import (
	"fmt"
	"testing"
)

func TestGobSerializer_Marshal(t *testing.T) {
	person := &Person{Name: "longerwu", Age: 23}
	serializer := GetSerializer(SerializationTypeGob)
	data, err := serializer.Marshal(person)
	fmt.Printf("data:%+v\n", string(data))
	if err != nil {
		t.Fatalf("Marshal err:%+v", err)
	}
	person2 := &Person{}
	if err := serializer.Unmarshal(data, person2); err != nil {
		t.Fatalf("Marshal err:%+v", err)
	}
	fmt.Println(person2)
}
