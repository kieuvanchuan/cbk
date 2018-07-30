package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"

	"../pack"
)

type User struct {
	PrivateKey *rsa.PrivateKey
	Add        [32]byte
	Balancer   int64
}

func CreatUser() User {
	// var rand io.Reader

	private, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		fmt.Println("create key for user error  :", err)
		return User{}
	}
	pub, err := json.Marshal(private.PublicKey)
	if err != nil {
		fmt.Println("Marshal(publickey) err : ")
		return User{}
	}
	ad := sha256.Sum256(pub)
	for i := 0; i < 32; i++ {
		if ad[i] < 65 || ad[i] > 122 || (90 < ad[i] && ad[i] < 97) {
			if 26 <= byte(math.Mod(float64(ad[i]), 52)) && byte(math.Mod(float64(ad[i]), 52)) <= 31 {
				ad[i] = byte(math.Mod(float64(ad[i]), 52)) + 65 + 10
			} else {
				ad[i] = byte(math.Mod(float64(ad[i]), 52)) + 65
			}
		}

	}
	usr := User{
		private,
		ad,
		0,
	}
	return usr
}

type args struct {
	Tranin []pack.Transaction
}
type tests struct {
	Name string
	Args args
	Want [32]byte
}

func main() {

	u1 := CreatUser()
	u2 := CreatUser()

	tran1 := pack.CrTran(u1.Add, u2.Add, 30, u1.PrivateKey)
	tran2 := pack.CrTran(u1.Add, u2.Add, 40, u1.PrivateKey)
	tran3 := pack.CrTran(u1.Add, u2.Add, 50, u1.PrivateKey)
	tran4 := pack.CrTran(u1.Add, u2.Add, 20, u1.PrivateKey)
	tran5 := pack.CrTran(u1.Add, u2.Add, 10, u1.PrivateKey)
	tran6 := pack.CrTran(u1.Add, u2.Add, 15, u1.PrivateKey)
	tran := []pack.Transaction{tran1, tran2, tran3}

	tranhash := make([][]byte, 5)
	for i := 0; i < 5; i++ {

		mid := bytes.Join([][]byte{tran.Tranin[i].SendAdd[0:32], tran.Tranin[i].ReceiveAdd[0:32], pack.Convert(tran.Tranin[i].Count)}, []byte{})
		tr := sha256.Sum256(mid)
		//tranhash[i] la gia tri bam cua transaction i
		tranhash[i] = tr[0:32]

	}

	tr := sha256.Sum256(bytes.Join(tranhash[0:2], []byte("")))
	copy(tranhash[0][:], tr[0:32])

	tr = sha256.Sum256(bytes.Join(tranhash[2:4], []byte("")))
	copy(tranhash[1][:], tr[0:32])

	tranhash = append(tranhash, tranhash[4])
	tr = sha256.Sum256(bytes.Join(tranhash[4:6], []byte("")))
	copy(tranhash[2][:], tr[0:32])

	tranhash = tranhash[:3]

	tranhash = append(tranhash, tranhash[2])
	tr = sha256.Sum256(bytes.Join(tranhash[0:2], []byte("")))
	copy(tranhash[0][:], tr[0:32])

	tr = sha256.Sum256(bytes.Join(tranhash[2:4], []byte("")))
	copy(tranhash[1][:], tr[0:32])

	fmt.Println(tranhash[0:2])
	tr = sha256.Sum256(bytes.Join(tranhash[0:2], []byte("")))
	copy(tranhash[0][:], tr[0:32])

	var root [32]byte
	copy(root[:], tranhash[0][0:32])

	test1 := []tests{
		{
			"test-1",
			tran,
			root,
		},
	}
	jsontest, _ := json.Marshal(test1)
	fmt.Println(string(jsontest))
	err := ioutil.WriteFile("test_mkroot.json", jsontest, 0644)
	fmt.Println(err)
	// var test []tests
	// file, err := os.Open("test_mkroot.json")

	// if err != nil {
	// 	fmt.Println("open genesis.json error: ", err)
	// 	return
	// }
	// defer file.Close()
	// jsonPar := json.NewDecoder(file)
	// jsonPar.Decode(&test)
	// fmt.Println(test)

}
