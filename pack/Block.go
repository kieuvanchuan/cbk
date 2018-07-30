package pack

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"time"
)

type Block struct {
	Time         int64
	Data         []Transaction
	Hash         [32]byte
	PreBlockHash [32]byte
	Nonce        int64
	MRoot        [32]byte
}

// chuyen kieu int64 sang []byte
func Convert(i641 int64) []byte {
	j := make([]byte, 8)
	for i := 0; i < 8; i++ {
		i64 := i641
		j[i] = byte(i64 << uint((i * 8)) >> 56)
	}
	return j
}

//tao block moi
func NewBlock(Nonce int64, Data []Transaction, PreBlockHash [32]byte, time int64) Block {
	var b Block
	b.Data = Data
	b.Nonce = Nonce
	b.PreBlockHash = PreBlockHash
	Hashtime := Convert(time)
	Mroot := MakeMRoot(b.Data)
	Head := bytes.Join([][]byte{Hashtime, Mroot[0:32], b.PreBlockHash[:], Convert(b.Nonce)}, []byte{})
	b.Hash = sha256.Sum256(Head)
	b.MRoot = Mroot
	return b
}

//hien thi thong tin block
func (b *Block) ShowInfor() {
	fmt.Println("block hash: ", b.Hash)
	fmt.Println("timestamp : ", time.Unix(b.Time, 0))
	fmt.Println("preblock hash : ", b.PreBlockHash)
	n := len(b.Data)

	for i := 0; i < n; i++ {
		sendAd := b.Data[i].SendAdd[0:32]
		receiveAd := b.Data[i].ReceiveAdd[0:32]

		fmt.Println(string(sendAd), " sen to ", string(receiveAd), " : ", b.Data[i].Count)
	}
}

// lay gia tri dac trung cua cac transaction
func MakeMRoot(tran []Transaction) [32]byte {
	n := len(tran)
	tranhash := make([][]byte, n)
	for i := 0; i < n; i++ {

		mid := bytes.Join([][]byte{tran[i].SendAdd[0:32], tran[i].ReceiveAdd[0:32], Convert(tran[i].Count)}, []byte{})
		tr := sha256.Sum256(mid)
		//tranhash[i] la gia tri bam cua transaction i
		tranhash[i] = tr[0:32]

	}
	// mid1 := bytes.Join(tranhash, []byte{})
	// Root := sha256.Sum256(mid1)

	for {
		if n == 1 {
			break
		}
		index := 0
		if n%2 != 0 {
			tranhash = append(tranhash[0:n], tranhash[n-1])
			n = n + 1
		}
		for i := 0; i < n; i = i + 2 {
			tr := sha256.Sum256(bytes.Join(tranhash[i:i+2], []byte("")))
			tranhash[index] = tr[:32]
			index++
		}
		tranhash = tranhash[:index]

		n = index

	}
	var Root [32]byte
	copy(Root[:], tranhash[0][:32])
	return Root

}

// kiem tra xem 1 block co phai block 0 hay khong
func CheckGennis(b Block) bool {
	for i := 0; i < 32; i++ {
		if uint8(b.Hash[i]) != 0 {
			return false
		}
	}
	return true
}

// kiem tra xem chuoi block co hop le hay khong
func CheckValid(chain []Block, topBlock Block) bool {
	n := len(chain)
	if n < 1 {

		return false
	}
	if chain[0].PreBlockHash != topBlock.Hash {
		return false
	}
	for i := 1; i < n; i++ {
		if (chain[i-1].Hash != chain[i].PreBlockHash) || !CheckBlock(chain[i]) {

			return false
		}

	}
	return true
}

// kiem tra xem block co hop le hay khong
func CheckBlock(b Block) bool {
	if b.MRoot != MakeMRoot(b.Data) {

		return false
	} else {
		Hashtime := b.Time
		Mroot := MakeMRoot(b.Data)
		Head := bytes.Join([][]byte{Convert(Hashtime), Mroot[0:32], b.PreBlockHash[:], Convert(b.Nonce)}, []byte{})
		if b.Hash != sha256.Sum256(Head) {

			return false
		}
	}
	return true
}
