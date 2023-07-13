package json

import (
	"testing"
)

func TestJsonUnmarshal(t *testing.T) {
	type User struct {
		Name string
		Age  int
		Sex  byte `name:"gender"`
	}

	type Book struct {
		ISBN     string `name:"isbn"`
		Name     string
		Price    float32  `name:"price"`
		Author   *User    `name:"author"`
		Keywords []string `name:"kws"`
		Local    map[int]bool
	}

	user := User{
		Name: "钱钟书",
		Age:  57,
		Sex:  1,
	}
	book := Book{
		ISBN:     "4243547567",
		Name:     "围城",
		Price:    34.8,
		Author:   &user,
		Keywords: []string{"爱情", "民国", "留学"},
		Local:    map[int]bool{2: true, 3: false},
	}
	var u User
	var b Book
	ustr, _ := JsonMarshal(user)
	bstr, _ := JsonMarshal(book)
	println(string(ustr))
	println(string(bstr))

	type args struct {
		data []byte
		v    any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test user struct",
			args: args{
				data: ustr,
				v:    &u,
			},
			wantErr: false,
		},
		{
			name: "test book struct",
			args: args{
				data: bstr,
				v:    &b,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := JsonUnmarshal(tt.args.data, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("JsonUnmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
