package codec

import (
	"reflect"
	"testing"
)

type Person struct {
	Name string `xml:"name"`
	Age  int    `xml:"age"`
}

func TestXmlSerializer_Marshal(t *testing.T) {
	type args struct {
		in interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantOut []byte
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				in: &Person{Age: 23, Name: "longerwu"},
			},
			wantOut: []byte(`<Person><name>longerwu</name><age>23</age></Person>`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := XmlSerializer{}
			gotOut, err := x.Marshal(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("Marshal() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
