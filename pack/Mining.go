package pack

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

const DIFF = 3

// minning block moi
func Mine(tran []Transaction, phash [32]byte) Block {
	var n int64 = 0
	Mroot := MakeMRoot(tran)
	time := time.Now().Unix()

	var Hash [32]byte
	for {
		Head := bytes.Join([][]byte{Convert(time), Mroot[0:32], phash[0:32], Convert(n)}, []byte{})
		Hash = sha256.Sum256(Head)
		if CheckPOW(Hash, DIFF) {
			block := Block{
				time,
				tran,
				Hash,
				phash,
				n,
				Mroot,
			}
			return block
		}
		n++

	}
}

//xac thuc transactionpool cha lai cac transaction hop le
func CheckTranPool(tran []Transaction, db *leveldb.DB) []Transaction {

	heigh, _ := db.Get([]byte("heigh"), nil)

	h := int64(binary.BigEndian.Uint64(heigh))
	var i int64
	chain := make([]Block, h+1)
	for i = 0; i < h+1; i++ {
		b, _ := db.Get(Convert(i), nil)

		c, _ := db.Get(b, nil)
		json.Unmarshal(c, &chain[i])

	}
	lenTran := len(tran)
	TranResult := make([]Transaction, 0)
	CoinIn := make([]int64, lenTran)
	CoinOut := make([]int64, lenTran)
	for i = 0; i < h+1; i++ {
		n := len(chain[i].Data)
		for j := 0; j < n; j++ {
			for k := 0; k < lenTran; k++ {
				if chain[i].Data[j].ReceiveAdd == tran[k].SendAdd {
					CoinIn[k] = CoinIn[k] + chain[i].Data[j].Count
				} else {
					if chain[i].Data[j].SendAdd == tran[k].SendAdd {
						CoinOut[k] = CoinOut[k] + chain[i].Data[j].Count
					}
				}
			}

		}
	}
	admine := []byte("FdKmobTRmXbKlfrcDdDQZHLsbmjhjQbs")
	var addr [32]byte
	copy(addr[:], admine[0:32])
	tranbase := Transaction{
		[32]byte{},
		addr,
		10,
		nil,
		rsa.PublicKey{},
	}
	TranResult = append(TranResult, tranbase)
	for k := 0; k < lenTran; k++ {
		if (CoinIn[k] - CoinOut[k]) > tran[k].Count {
			TranResult = append(TranResult, tran[k])
		}
	}
	return TranResult

}

// kiem tra tinh hop le cua gia tri bam voi do kho cua thuat toan POW
func CheckPOW(b [32]byte, diff int) bool {
	for i := 0; i < diff; i++ {
		if uint8(b[i]) != 0 {
			return false
		}
	}
	return true

}

// kiem tra tinh hop le cua mot transaction
func CheckTran(tran Transaction, db *leveldb.DB) bool {
	mid := bytes.Join([][]byte{tran.SendAdd[0:32], tran.ReceiveAdd[0:32], Convert(tran.Count)}, []byte{})
	hash := sha256.Sum256(mid)
	pub, err := json.Marshal(tran.Publickey)
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
	if ad != tran.SendAdd {
		return false
	}
	err = rsa.VerifyPSS(&tran.Publickey, crypto.SHA256, hash[:], tran.Signature, nil)
	if err != nil {
		return false
	}

	heigh, _ := db.Get([]byte("heigh"), nil)

	h := int64(binary.BigEndian.Uint64(heigh))
	var i int64
	chain := make([]Block, h+1)
	for i = 0; i < h+1; i++ {
		b, err := db.Get(Convert(i), nil)
		if err != nil {
			fmt.Println("loi doc du lieu tu db minning--> CheckTran() :", err)
			return false
		}
		c, err := db.Get(b, nil)
		if err != nil {
			fmt.Println("loi doc du lieu tu db minning--> CheckTran() :", err)
			return false
		}
		json.Unmarshal(c, &chain[i])

	}
	CoinIn := int64(0)
	CoinOut := int64(0)
	for i = 0; i < h+1; i++ {
		n := len(chain[i].Data)
		for j := 0; j < n; j++ {

			if chain[i].Data[j].ReceiveAdd == tran.SendAdd {

				CoinIn = CoinIn + chain[i].Data[j].Count
			} else {
				if chain[i].Data[j].SendAdd == tran.SendAdd {
					CoinOut = CoinOut + chain[i].Data[j].Count
				}
			}

		}
	}

	if (CoinIn - CoinOut) > tran.Count {
		return true
	} else {
		return false
	}

}
