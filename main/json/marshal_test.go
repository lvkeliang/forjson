package json

import (
	"fmt"
	"reflect"
	"testing"
)

func TestJsonMarshal(t *testing.T) {
	type JsonTest struct {
		MapTest    map[int]string
		IntTest    int
		StringTest string
		SliceTest  []string `name:"sliceTest"`
	}
	var a = make(map[int]string)
	a[1] = "1"
	a[2] = "3"
	testStruct := JsonTest{
		MapTest:    a,
		IntTest:    65535,
		StringTest: "test",
		SliceTest:  []string{"test1", "test2", "test3"},
	}
	jsonStr := []byte("{\"MapTest\":{\"1\":\"1\",\"2\":\"3\"},\"IntTest\":65535,\"StringTest\":\"test\",\"sliceTest\":[\"test1\",\"test2\",\"test3\"]}")

	type args struct {
		obj any
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "testMarshal",
			args: args{
				obj: testStruct,
			},
			want:    jsonStr,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JsonMarshal(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("JsonMarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonMarshal() got = %v,\n want %v", got, tt.want)
			}
		})
	}
	res, err := JsonMarshal(testStruct)
	fmt.Printf("res: %v\nerr: %v\n", string(res), err)
	fmt.Printf("want: %v\n", string(jsonStr))
}
