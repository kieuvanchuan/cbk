package pack

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestConvert(t *testing.T) {
	type args struct {
		i641 int64
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
		{
			"test 1",
			args{
				int64(256),
			},
			[]byte{0, 0, 0, 0, 0, 0, 1, 0},
		},
		{
			"test 2",
			args{
				int64(255),
			},
			[]byte{0, 0, 0, 0, 0, 0, 0, 255},
		}, {
			"test 3",
			args{
				int64(257),
			},
			[]byte{0, 0, 0, 0, 0, 0, 1, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Convert(tt.args.i641); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeMRoot(t *testing.T) {
	type args struct {
		Tranin []Transaction
	}
	type test struct {
		Name string
		Args args
		Want [32]byte
	}
	var tests []test
	file, err := os.Open("../Test/test_mkroot.json")

	if err != nil {
		fmt.Println("open genesis.json error: ", err)
		return
	}
	defer file.Close()
	jsonPar := json.NewDecoder(file)
	jsonPar.Decode(&tests)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := MakeMRoot(tt.Args.Tranin); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("MakeMRoot() = %v, want %v", got, tt.Want)
			}
		})
	}
}
