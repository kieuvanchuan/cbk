package main

import (
	"fmt"
	"../pack"

	"github.com/syndtr/goleveldb/leveldb"
)

func main() {

	N := pack.CreateNode("localhost:7777")
	db, err := leveldb.OpenFile("./Data/Block.db", nil)
	fmt.Println(err)
	h, err := db.Get([]byte("heigh"), nil)

	if err != nil && h == nil {
		pack.CreateBlockGenis("genesis.json", db)
	}
	go N.RunMiner(db)
	N.RunNode(db)

}
