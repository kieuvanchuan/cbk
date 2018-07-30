package pack

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

type Node struct {
	Add      string
	NextNode []string
	TranPool []Transaction
	HashTran [][32]byte
}

func CreateNode(add string) Node {
	var node Node = Node{
		add,
		[]string{},
		[]Transaction{},
		[][32]byte{},
	}
	return node
}

type NexNode struct {
	S []string
}

func (n *Node) AddNextNode(add string, db *leveldb.DB) {
	conn, err := n.Connect(add)
	if err != nil {
		fmt.Println("Node khong kha dung ")
		return
	}
	conn.Close()
	lenAdd := len(n.NextNode)
	for i := 0; i < lenAdd; i++ {
		if add == n.NextNode[i] {
			fmt.Println("node existed ")
			return
		}
	}

	n.NextNode = append(n.NextNode, add)
	s := NexNode{
		n.NextNode,
	}

	marshals, err := json.Marshal(s)
	if err != nil {
		fmt.Println("erro marshal in function addnode :", err)
		n.NextNode = nil
		return
	}
	// fmt.Println(s)
	err = db.Put([]byte("nextnode"), marshals, nil)
	if err != nil {
		fmt.Println("write nextnode err", err)
		return
	}

}

// connect toi node 1

func (N *Node) Connect(add string) (*net.TCPConn, error) {
	tcpAd, err := net.ResolveTCPAddr("tcp4", add)
	if nil != err {
		return nil, err
	} else {
		conn, err := net.DialTCP("tcp", nil, tcpAd)
		if err != nil {
			return nil, err
		} else {
			return conn, err
		}
	}
}

//mo ket noi voi cac node khac
func (N *Node) Listen(db *leveldb.DB) error {
	tcpAd, err := net.ResolveTCPAddr("tcp4", N.Add)
	if err != nil {
		return err
	} else {
		listen, err := net.ListenTCP("tcp", tcpAd)
		if nil != err {
			return err
		} else {

			for {
				conn, err := listen.Accept()

				if nil != err {
					continue
				} else {

					go N.handle(conn, db)
				}
			}

		}

	}

}

// ham xu ly khi co mot connect gui mot message den
func (N *Node) handle(conn net.Conn, db *leveldb.DB) {

	b := make([]byte, 2000000)
	Mess := Message{}

	for {

		count, err := conn.Read(b)
		if err != nil {

			conn.Close()
			return
		}
		// fmt.Println(b)

		err = json.Unmarshal(b[:count], &Mess)

		if nil != err {
			continue
		} else {
			break
		}
	}

	// reader := bufio.NewReader(os.Stdin)
	// fmt.Println("enter path data file : ")
	// text, _ := reader.ReadString('\n')

	switch Mess.Cmd {
	case GetHeighBlock:
		fmt.Println("nhan duoc yeu cau gui heigh block")
		heigh, err := db.Get([]byte("heigh"), nil)
		if nil != err {

			conn.Write(Convert(int64(0)))

			break
		} else {

			conn.Write(heigh)

			break
		}
	case SendBlock: // nhan duoc mess co cmd = senblock check cac block co hop le khong
		// sau do ghi cac block vao data neu cac block hop lep
		// Data := Mess.Data
		fmt.Println("block duoc gui den ")
		handleSendBlock(Mess, db, conn)
		break
	case DistributeBlock: // nhan duoc mess gui block moi duoc tao den
		// check tinh dung dan cua block va them vao db
		fmt.Println("nhan duoc block moi")
		handleDistributeBlock(Mess, db, conn, N)
		break
	case GetBlock: // nhan duoc thong tin co cm la getblock
		// kiem tra xem do cao cua blockchain cua node hien tai cao hon chieu cao chain
		//cua khoi yeu cau block hay khong, cao hon thi gui cac block ma node yeu cau chua co
		fmt.Println("nhan duoc yeu cau gui block")
		handleGetBlock(Mess, db, conn)
		break

	case CreateTran:
		fmt.Println("nhan duoc transaction moi")
		handleCreateTran(N, Mess, db, conn)
		break

	case GetBalancer:
		fmt.Println("wallet getbalencer")
		acc := getBalencer(Mess.Acc, db, conn)
		mss := Message{
			GetBalancer,
			0,
			nil,
			nil,
			acc,
		}
		mes, err := json.Marshal(mss)
		if err != nil {
			fmt.Println("can't Marshal json :  ", err)
			conn.Write([]byte("errr"))
			break
		}
		conn.Write(mes)
		break
	default:
		fmt.Println("---------")
		break
	}
	conn.Close()
	fmt.Print("> ")

}

// lay so du cua danh sach cac acc dau vao va cha lai danh sach account voi so du
func getBalencer(acc []Account, db *leveldb.DB, conn net.Conn) []Account {
	n := len(acc)

	heigh, err := db.Get([]byte("heigh"), nil)
	if err != nil {
		fmt.Println("can't read database :  ", err)
		return []Account{}
	}
	h := int64(binary.BigEndian.Uint64(heigh))
	var i int64
	chain := make([]Block, h+1)
	for i = 0; i < h+1; i++ {
		b, err := db.Get(Convert(i), nil)
		if err != nil {
			fmt.Println("can't read database :  ", err)
			return []Account{}
		}
		c, err := db.Get(b, nil)
		if err != nil {
			fmt.Println("can't read database :  ", err)
			return []Account{}
		}
		json.Unmarshal(c, &chain[i])

	}
	CoinIn := make([]int64, n)
	CoinOut := make([]int64, n)
	for i = 0; i < h+1; i++ {
		l := len(chain[i].Data)
		for j := 0; j < l; j++ {
			for k := 0; k < n; k++ {
				if chain[i].Data[j].ReceiveAdd == acc[k].Add {

					CoinIn[k] = CoinIn[k] + chain[i].Data[j].Count
				} else {
					if chain[i].Data[j].SendAdd == acc[k].Add {
						CoinOut[k] = CoinOut[k] + chain[i].Data[j].Count
					}
				}
			}
		}
	}
	for i := 0; i < n; i++ {
		acc[i].Balancer = CoinIn[i] - CoinOut[i]
	}
	return acc
}

// gui yeu cau lay block tu node khac
func (n *Node) GetBlock(conn *net.TCPConn, db *leveldb.DB) {

	heigh, err := db.Get([]byte("heigh"), nil)
	if err != nil {
		fmt.Println("can't read database :  ", err)
		return
	}
	h := int64(binary.BigEndian.Uint64(heigh))

	mess := Message{
		GetBlock,
		h,
		nil,
		nil,
		nil,
	}
	ms, err := json.Marshal(mess)
	if err != nil {
		fmt.Println("can't Marshal Message :  ", err)
		return
	}
	conn.Write(ms)
}

//lay do dai chuoi cua node duoc ket noi bang bien conn

func (n *Node) GetHeighBlock(conn *net.TCPConn) {
	mess := Message{
		GetHeighBlock,
		0,
		nil,
		nil,
		nil,
	}
	ms, err := json.Marshal(mess)
	if err != nil {
		fmt.Println("can't Marshal Message :  ", err)
		return
	}
	conn.Write(ms)
	b := make([]byte, 8)
	_, err = conn.Read(b)
	if err != nil {
		fmt.Println("can't read data from connect :  ", err)
		return
	}
	fmt.Println("heigh : ", int64(binary.BigEndian.Uint64(b)))
	conn.Close()
}

//gui block moi toi tat ca cac node co the ket noi duoc trong danh sach nextnode
func (n *Node) DistributeBlock(b Block, h int64) {
	l := len(n.NextNode)
	block := []Block{}
	block = append(block, b)
	mess := Message{
		DistributeBlock,
		h,
		block,
		nil,
		nil,
	}
	mss, err := json.Marshal(mess)
	if err != nil {
		fmt.Println("can't Marshal in distributeblock() :  ", err)
		return
	}

	for i := 0; i < l; i++ {
		conn, err := n.Connect(n.NextNode[i])

		if nil != err {
			continue
		} else {
			conn.Write(mss)
		}
		defer conn.Close()
	}
}

func Check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// create block so 0 voi file genesis.json
func CreateBlockGenis(s string, db *leveldb.DB) {
	var B Block
	file, err := os.Open(s)

	if err != nil {
		fmt.Println("open genesis.json error: ", err)
		return
	}
	defer file.Close()
	jsonPar := json.NewDecoder(file)
	jsonPar.Decode(&B)
	h := Convert(int64(0))
	err = db.Put(h, B.Hash[:], nil)
	if err != nil {
		fmt.Println("can't write data Createblock fail : ", err)
		return
	}
	db.Put([]byte("heigh"), h, nil)
	value, err := json.Marshal(B)
	if err != nil {
		fmt.Println("can't Marshal Block in creategenesisblock() : ", err)
		return
	}
	err = db.Put(B.Hash[:], value, nil)
	if err != nil {
		fmt.Println("can't write data Createblock fail : ", err)
		return
	}

}

//doc du lieu vao tu console va phan tich thanh cau lenh va cac tham so
func ReadFunc() []string {
	reader := bufio.NewReader(os.Stdin)
	var text string
	var err error
	for {
		text, err = reader.ReadString('\n')

		if err != nil {
			fmt.Println("nhap lai lenh")
			continue
		} else {
			break
		}
	}
	text = strings.TrimSpace(text)
	s1 := strings.Split(text, "(")

	if len(s1) >= 2 {
		s2 := s1[1]
		s2 = strings.TrimSpace(s2)
		s2 = strings.Replace(s2, ")", "", 1)
		s3 := strings.Split(s2, ",")
		s := make([]string, 0)
		s = append(s, s1[0])

		for i := 0; i < len(s3); i++ {
			s = append(s, strings.TrimSpace(s3[i]))
		}

		return s
	}
	return nil
}

// gui transaction toi cac node co the ket noi khi co transaction moi duoc tao
func (N *Node) sendTran(tran Transaction) {
	fmt.Println("send tran")
	n := len(N.NextNode)
	if n > 0 {

		for i := 0; i < n; i++ {
			conn1, err := N.Connect(N.NextNode[i])

			if err != nil {
				continue
			} else {
				Tr := []Transaction{}
				Tr = append(Tr, tran)
				mess := Message{
					CreateTran,
					0,
					nil,
					Tr,
					nil,
				}
				mss, err := json.Marshal(mess)
				if err != nil {
					fmt.Println("can't Marshal Message in Sentran() : ", err)
					return
				}

				conn1.Write(mss)

			}
			conn1.Close()
		}
	}
}

// kiem tra xem transaction da ton tai trong transactionpool chua
func CheckHash(hash [32]byte, hashpool [][32]byte) bool {
	n := len(hashpool)

	for i := 0; i < n; i++ {

		if hash == hashpool[i] {

			return false
		}

	}
	return true
}

// ham xu ly khi message den co command la SendBlock()
func handleSendBlock(Mess Message, db *leveldb.DB, conn net.Conn) {
	h, err := db.Get([]byte("heigh"), nil)
	if err != nil {
		fmt.Println("can't get data from database : ", err)
		return
	}
	hash, err := db.Get(h, nil)
	if err != nil {
		fmt.Println("can't get data from database : ", err)
		return
	}
	B, err := db.Get(hash, nil)
	if err != nil {
		fmt.Println("can't get data from database : ", err)
		return
	}
	Bl := Block{}
	err = json.Unmarshal(B, &Bl)
	if err != nil {
		fmt.Println("can't Unmarshal json : ", err)
		return
	}
	if CheckValid(Mess.Data, Bl) && Mess.Heigh > 0 {

		n := len(Mess.Data)
		heigh, err := db.Get([]byte("heigh"), nil)
		if err != nil {

			if CheckGennis(Mess.Data[0]) {
				var i int64
				for i = 0; i < int64(n); i++ {
					Data1, err := json.Marshal(Mess.Data[i])
					if err != nil {
						fmt.Println("can't Marshal Message  : ", err)
						return
					}
					db.Put(Convert(i), Mess.Data[i].Hash[:], nil)
					db.Put(Mess.Data[i].Hash[:], Data1, nil)

					return
				}
			} else {
				fmt.Println("Block not Valid ")
				return
			}
		} else {
			// h := int64(binary.BigEndian.Uint64(heigh)
			BlockHash, err := db.Get(heigh, nil)
			if err != nil {
				fmt.Println("can't get data from database : ", err)
				return
			}
			BlockJson, err := db.Get(BlockHash, nil)
			var TopBlock Block
			json.Unmarshal(BlockJson, &TopBlock)

			if TopBlock.Hash != Mess.Data[0].PreBlockHash {

				fmt.Println("Block data error")

				return

			} else {
				h := int64(binary.BigEndian.Uint64(heigh))

				var i int64
				for i = 0; i < int64(n); i++ {
					h = h + 1
					Data, err := json.Marshal(Mess.Data[i])
					if err != nil {
						fmt.Println("can't Marshal Message  : ", err)
						return
					}

					db.Put(Convert(h), Mess.Data[i].Hash[:], nil)
					db.Put(Mess.Data[i].Hash[:], Data, nil)

				}

				err := db.Put([]byte("heigh"), Convert(h), nil)
				if err != nil {
					fmt.Println("can't write data : ", err)
					return
				}

				return

			}

		}
	} else {
		fmt.Println("node has lesser or equal block", Mess.Heigh)
		return

	}
}

// xu ly message co command la distributeBlock
func handleDistributeBlock(Mess Message, db *leveldb.DB, conn net.Conn, N *Node) {
	heigh, err := db.Get([]byte("heigh"), nil)
	if err != nil {
		fmt.Println("can't get data from database : ", err)
		return
	}
	top, err := db.Get(heigh, nil)
	if err != nil {
		fmt.Println("can't get data from database : ", err)
		return
	}
	topB, err := db.Get(top, nil)
	if err != nil {
		fmt.Println("can't get data from database : ", err)
		return
	}
	var topBlock Block
	json.Unmarshal(topB, &topBlock)
	if topBlock.Hash != Mess.Data[0].PreBlockHash {
		fmt.Println("yeu cau cap nhat blockchain")
		return
	} else {
		if CheckBlock(Mess.Data[0]) {
			h := int64(binary.BigEndian.Uint64(heigh))
			h = h + 1
			Block1, err := json.Marshal(Mess.Data[0])
			if err != nil {
				fmt.Println("can't Marshal Message : ", err)
				return
			}

			db.Put(Convert(h), Mess.Data[0].Hash[:], nil)
			db.Put(Mess.Data[0].Hash[:], Block1, nil)
			db.Put([]byte("heigh"), Convert(h), nil)
			N.TranPool = []Transaction{}
			N.HashTran = nil
			fmt.Println("imported new block")
			return
		} else {
			fmt.Println("Block data error")
			return
		}
	}
}

//xu ly message co command la GetBlock
func handleGetBlock(Mess Message, db *leveldb.DB, conn net.Conn) {
	heigh, err := db.Get([]byte("heigh"), nil)
	if err != nil {
		fmt.Println("can't get data from database : ", err)
		return
	}
	h := int64(binary.BigEndian.Uint64(heigh))
	if h <= Mess.Heigh {
		Mss := Message{
			SendBlock,
			0,
			nil,
			nil,
			nil,
		}
		Mess1, _ := json.Marshal(Mss)
		conn.Write(Mess1)
		return
	} else {

		chain := make([]Block, h-Mess.Heigh)
		for i := Mess.Heigh + 1; i <= h; i++ {
			b, err := db.Get(Convert(i), nil)
			if err != nil {
				fmt.Println("can't get data from database : ", err)
				return
			}
			c, err := db.Get(b, nil)
			if err != nil {
				fmt.Println("can't get data from database : ", err)
				return
			}
			var bl Block = Block{}
			err = json.Unmarshal(c, &bl)
			if err != nil {
				fmt.Println("can't UnMarshal Message : ", err)
				return
			}
			chain[i-Mess.Heigh-1] = bl

		}
		Mss := Message{
			SendBlock,
			h - Mess.Heigh,
			chain,
			nil,
			nil,
		}

		Mess1, err := json.Marshal(Mss)
		if err != nil {
			fmt.Println("can't Marshal Message : ", err)
			return
		}
		conn.Write(Mess1)

		fmt.Println("sended blockchain")
		// fmt.Println("-----", conn.RemoteAddr(), "----", conn.LocalAddr())
		return
	}
}

//xu ly message co command la Createtran
func handleCreateTran(N *Node, Mess Message, db *leveldb.DB, conn net.Conn) {
	if CheckTran(Mess.Tran[0], db) {

		if len(N.HashTran) == 0 {
			mss := Transaction{
				Mess.Tran[0].SendAdd,
				Mess.Tran[0].ReceiveAdd,
				Mess.Tran[0].Count,
				nil,
				Mess.Tran[0].Publickey,
			}
			bhash, err := json.Marshal(mss)
			if err != nil {
				fmt.Println("can't Marshal Message in handlecreateTran() : ", err)
				return
			}
			hash := sha256.Sum256(bhash)
			N.HashTran = append(N.HashTran, hash)
			N.TranPool = append(N.TranPool, Mess.Tran[0])
			mess := Message{
				"transaction tao thanh công ",
				0,
				nil,
				nil,
				nil,
			}
			mss1, err := json.Marshal(mess)
			if err != nil {
				fmt.Println("can't Marshal Message : ", err)
				return
			}

			conn.Write(mss1)
			N.sendTran(Mess.Tran[0])
			return
		} else {
			mss := Transaction{
				Mess.Tran[0].SendAdd,
				Mess.Tran[0].ReceiveAdd,
				Mess.Tran[0].Count,
				nil,
				Mess.Tran[0].Publickey,
			}
			bhash, err := json.Marshal(mss)
			if err != nil {
				fmt.Println("can't Marshal Message : ", err)
				return
			}
			hash := sha256.Sum256(bhash)
			if CheckHash(hash, N.HashTran) {
				N.HashTran = append(N.HashTran, hash)
				N.TranPool = append(N.TranPool, Mess.Tran[0])
				mess := Message{
					"transaction tao thanh công ",
					0,
					nil,
					nil,
					nil,
				}
				mss1, err := json.Marshal(mess)
				if err != nil {
					fmt.Println("can't Marshal Message : ", err)
					return
				}

				conn.Write(mss1)
				N.sendTran(Mess.Tran[0])

				return
			} else {
				mess := Message{
					"transaction khong hop le",
					0,
					nil,
					nil,
					nil,
				}
				mss, err := json.Marshal(mess)
				if err != nil {
					fmt.Println("can't Marshal Message : ", err)
					return
				}

				conn.Write(mss)
				return
			}
		}
	} else {
		mess := Message{
			"transaction khong hop le",
			0,
			nil,
			nil,
			nil,
		}
		mss, err := json.Marshal(mess)
		if err != nil {
			fmt.Println("can't Marshal Message : ", err)
			return
		}

		conn.Write(mss)
		return
	}
}

// khoi chay mot node
func (N *Node) RunNode(db *leveldb.DB) {
	nextnd := NexNode{}
	nextndjson, err := db.Get([]byte("nextnode"), nil)
	if err == nil {
		err = json.Unmarshal(nextndjson, &nextnd)
		if err != nil {
			fmt.Println("can't get lis nextnode")
		} else {
			N.NextNode = nextnd.S
		}
	}
	go N.Listen(db)

	fmt.Println("addnode(add string) : them mot node vao cac node co the connect")
	fmt.Println("getblock(add string) : yeu cau lay block tu node co dia chi add")
	fmt.Println("getheighblock(add string) : lay chieu dai blockchian cua node co dia chi add")
	fmt.Println("heighblock() : in ra chieu dai blockchain cua node")
	fmt.Println("showinforblock(int) : show thong tin block")
	fmt.Println("showtranpool() : hien thi thong tin pool")
	fmt.Println("exit() : thoat chuong trinh")
	fmt.Println("showlistnode() : hien thi danh sach node co the connect")
	for {
		fmt.Print("\n> ")
		command := ReadFunc()
		if command == nil {
			fmt.Println("not found command ")
			continue
		}
		n := len(command)
		switch strings.ToLower(strings.TrimSpace(command[0])) {
		case "addnode":
			if n != 2 {
				fmt.Println("input command error ")
				break
			} else {
				N.AddNextNode(command[1], db)
				break
			}
		case "getblock":
			if n != 2 {
				fmt.Println("input command error")
				break
			} else {
				conn, err := N.Connect(command[1])
				if err != nil {
					fmt.Println("can't connect to node : ", err)
					break
				}
				N.GetBlock(conn, db)

				N.handle(conn, db)

				break
			}
		case "getheighblock":
			if n != 2 {
				fmt.Println("input command error")
				break
			} else {
				conn, err := N.Connect(command[1])

				if err != nil {
					fmt.Println("can't connect to node  : ", err)
					break
				}
				N.GetHeighBlock(conn)

				break
			}
		case "heighblock":

			heigh, err := db.Get([]byte("heigh"), nil)
			if err != nil {
				fmt.Println("can't get data from database : ", err)
				break
			}
			h := int64(binary.BigEndian.Uint64(heigh))
			fmt.Println(h)
			break
		case "showinforblock":
			h, err := strconv.Atoi(command[1])
			if err != nil {
				fmt.Println("can't convert string to integer : ", err)
				break
			}
			hash, err := db.Get(Convert(int64(h)), nil)
			if err != nil {
				fmt.Println("can't get data from database : ", err)
				break
			}
			B, err := db.Get(hash, nil)
			if err != nil {
				fmt.Println("can't get data from database : ", err)
				break
			}
			Bl := Block{}
			err = json.Unmarshal(B, &Bl)
			if err != nil {
				fmt.Println("can't Unmarshal json : ", err)
				break
			}
			Bl.ShowInfor()
			break
		case "showtranpool":
			fmt.Println("len(tranpool): ", len(N.TranPool))
			fmt.Println(N.TranPool)
			break
		case "showlistnode":
			fmt.Println(N.NextNode)
			break
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("command wrong")
			break
		}

	}

}

// kiem tra dieu kien tranpool va mining block moi
func (N *Node) RunMiner(db *leveldb.DB) {

	for {

		if len(N.TranPool) >= 3 {
			Tran := CheckTranPool(N.TranPool, db)
			N.TranPool = nil
			h, err := db.Get([]byte("heigh"), nil)

			if err != nil {
				fmt.Println("can't get data from database : ", err)
				continue
			}
			hs, err := db.Get(h, nil)
			var hash [32]byte
			copy(hash[:], hs[0:32])
			fmt.Println("minning")
			B := Mine(Tran, hash)
			h1 := int64(binary.BigEndian.Uint64(h)) + 1
			err = db.Put([]byte("heigh"), Convert(h1), nil)
			if err != nil {
				fmt.Println("can't write data to database : ", err)
				continue
			}
			err = db.Put(Convert(h1), B.Hash[:], nil)
			if err != nil {
				fmt.Println("can't write data to database : ", err)
				continue
			}
			bl, err := json.Marshal(B)
			if err != nil {
				fmt.Println("can't Marshal Block : ", err)
				continue
			}
			err = db.Put(B.Hash[:], bl, nil)
			if err != nil {
				fmt.Println("can't write data : ", err)
				continue
			}
			fmt.Println("minned")
			fmt.Println("> ")
			N.DistributeBlock(B, int64(binary.BigEndian.Uint64(h)+1))
		}
		time.Sleep(20 * time.Second)

	}

}
