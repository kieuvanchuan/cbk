package pack

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestCheckHash(t *testing.T) {
	type args struct {
		hash     [32]byte
		hashpool [][32]byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckHash(tt.args.hash, tt.args.hashpool); got != tt.want {
				t.Errorf("CheckHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_sendTran(t *testing.T) {
	type fields struct {
		Add      string
		NextNode []string
		TranPool []Transaction
		HashTran [][32]byte
	}
	type args struct {
		tran Transaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			N := &Node{
				Add:      tt.fields.Add,
				NextNode: tt.fields.NextNode,
				TranPool: tt.fields.TranPool,
				HashTran: tt.fields.HashTran,
			}
			N.sendTran(tt.args.tran)
		})
	}
}

func Test_handleSendBlock(t *testing.T) {
	type args struct {
		Mess Message
		db   *leveldb.DB
		conn net.Conn
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleSendBlock(tt.args.Mess, tt.args.db, tt.args.conn)
		})
	}
}

func Test_handleGetBlock(t *testing.T) {
	type args struct {
		Mess Message
		db   *leveldb.DB
		conn net.Conn
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleGetBlock(tt.args.Mess, tt.args.db, tt.args.conn)
		})
	}
}

func TestReadFunc(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		// TODO: Add test cases.
		{
			"test 1",
			[]string{"func1", "x1", "x2", "x3", "x4", "x5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempfile, err := ioutil.TempFile("", "TestFunRead")
			if err != nil {
				fmt.Println("Test err : ", err)
				return

			}
			defer os.Remove(tempfile.Name())
			conten := []byte("func1(   x1,   x2,  x3,  x4 , x5 ) \n")
			if _, err := tempfile.Write(conten); err != nil {
				fmt.Println("test err : ", err)
				return
			}

			if _, err := tempfile.Seek(0, 0); err != nil {
				fmt.Println("test err : ", err)
				return
			}
			oldstdin := os.Stdin
			defer func() { os.Stdin = oldstdin }()
			os.Stdin = tempfile

			got := ReadFunc()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadFunc() = %v, want %v", got, tt.want)
			}
			if err := tempfile.Close(); err != nil {
				fmt.Println("test err : ", err)
				return
			}
		})
	}
}
