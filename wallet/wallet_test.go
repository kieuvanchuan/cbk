package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"sync"
	"testing"

	"../pack"
)

func TestCreatTran(t *testing.T) {
	type args struct {
		SendAd    [32]byte
		ReceiveAd [32]byte
		Count     int64
		PrKey     *rsa.PrivateKey
	}
	type tt struct {
		Name string
		Args args
		Want pack.Transaction
	}
	var tests []tt

	file, err := os.Open("../Test/test_createtran.json")

	if err != nil {
		fmt.Println("open genesis.json error: ", err)
		return
	}
	defer file.Close()
	jsonPar := json.NewDecoder(file)
	jsonPar.Decode(&tests)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := CreatTran(tt.Args.SendAd, tt.Args.ReceiveAd, tt.Args.Count, tt.Args.PrKey)
			tt.Want.Signature = got.Signature
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("CreatTran() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestSendTran(t *testing.T) {

	type args struct {
		Tranin pack.Transaction
	}
	type test struct {
		Name string
		Args args
	}
	var tests []test

	file, err := os.Open("../Test/test_sendtran.json")

	if err != nil {
		fmt.Println("open genesis.json error: ", err)
		return
	}
	defer file.Close()
	jsonPar := json.NewDecoder(file)
	jsonPar.Decode(&tests)

	for _, tt := range tests {
		var mess pack.Message
		var tran pack.Transaction
		var wt sync.WaitGroup
		wt.Add(1)
		go func() {
			for {
				tcpAdd, err := net.ResolveTCPAddr("tcp4", "localhost:3333")
				if err != nil {
					continue
				}
				listen, err := net.ListenTCP("tcp", tcpAdd)
				if err != nil {
					continue
				}
				conn, err := listen.Accept()
				b := make([]byte, 200000)
				n, _ := conn.Read(b)

				json.Unmarshal(b[:n], &mess)

				tran = mess.Tran[0]
				wt.Done()
				conn.Close()
				return
			}
		}()
		tcpAd, _ := net.ResolveTCPAddr("tcp4", "localhost:3333")
		conn, _ := net.DialTCP("tcp", nil, tcpAd)
		defer conn.Close()
		t.Run(tt.Name, func(t *testing.T) {
			SendTran(tt.Args.Tranin, conn)

			wt.Wait()

			if !reflect.DeepEqual(tran, tt.Args.Tranin) {
				t.Errorf("SendTran() = %v, want %v", tt.Args.Tranin, tran)
			}
		})
	}
}
