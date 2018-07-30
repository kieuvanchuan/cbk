package pack

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

type Transaction struct {
	SendAdd    [32]byte
	ReceiveAdd [32]byte
	Count      int64
	Signature  []byte
	Publickey  rsa.PublicKey
}

func CrTran(send [32]byte, receive [32]byte, count int64, privatekey *rsa.PrivateKey) Transaction {
	mid := bytes.Join([][]byte{send[0:32], receive[0:32], Convert(count)}, []byte{})
	hash := sha256.Sum256(mid)
	sign, err := rsa.SignPSS(rand.Reader, privatekey, crypto.SHA256, hash[:], nil)
	Check(err)
	tran := Transaction{
		send,
		receive,
		count,
		sign,
		privatekey.PublicKey,
	}
	return tran
}

func CheckBaseTran(tran Transaction) {

}
