package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"../pack"

	"github.com/syndtr/goleveldb/leveldb"
)

type User struct {
	PrivateKey *rsa.PrivateKey
	Add        [32]byte
	Balancer   int64
}

// tao mot user moi
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

// connect toi node co dia chi la s
func ConnectNode(s string) (*net.TCPConn, error) {
	tcpAd, err := net.ResolveTCPAddr("tcp4", s)
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		conn, err := net.DialTCP("tcp", nil, tcpAd)
		if err != nil {
			fmt.Println(err)
			return nil, err

		} else {
			return conn, err
		}
	}
}

// tao mot transaction moi. cha lai transaction duoc tao
func CreatTran(sendAd [32]byte, receiveAd [32]byte, count int64, PrKey *rsa.PrivateKey) pack.Transaction {
	return pack.CrTran(sendAd, receiveAd, count, PrKey)
}

//gui transaction len node duoc kket noi boi bien conn
func SendTran(Tranin pack.Transaction, conn *net.TCPConn) bool {
	tran := []pack.Transaction{}
	tran = append(tran, Tranin)
	mess := pack.Message{
		pack.CreateTran,
		int64(0),
		nil,
		tran,
		nil,
	}
	b, err := json.Marshal(mess)
	if err != nil {
		fmt.Println(" Marshal Message fail in func SendTran() :", err)
		return false
	}

	_, err = conn.Write(b)

	if err != nil {
		conn.Close()
		return false
	} else {
		res := make([]byte, 20000)
		n, err1 := conn.Read(res)
		if err1 != nil {
			fmt.Println("response from node err :", err)
			return false
		}
		mss := pack.Message{}
		err := json.Unmarshal(res[:n], &mss)
		if err != nil {
			fmt.Println("Unmarshal response from node fail :", err)
			return false
		}
		fmt.Println(mss.Cmd)
		conn.Close()
		return true
	}

}

// hien thi danh sach cac user hien co trong wallet
func showLisUsr(file *leveldb.DB) {

	count, err := file.Get([]byte("count"), nil)
	if err != nil {
		fmt.Println("can't read database :", err)
		return
	}
	h := int64(binary.BigEndian.Uint64(count)) + 1
	lis := []User{}
	for i := int64(0); i < h; i++ {
		u := User{}

		b0, err := file.Get(pack.Convert(i), nil)
		if err != nil {
			fmt.Println("can't read database :", err)
			return
		}

		b, err := file.Get(b0, nil)
		if err != nil {
			fmt.Println("can't read database :", err)
			return
		}
		err = json.Unmarshal(b, &u)
		if err != nil {
			fmt.Println("can't read database :", err)
			return
		}
		lis = append(lis, u)
		fmt.Println(i+1, "-----", " add: ", string(u.Add[:]), "-----", "Balancer :", u.Balancer)

	}

}
func main() {
	NodeAdd := make([]string, 0)
	dbusr, _ := leveldb.OpenFile("./Data/Wallet.db", nil)

	fmt.Println("command createuser(): tao user moi")
	fmt.Println("command connectnode( add of node) : connect toi 1 node ")
	fmt.Println("command createtran(int(stt cua user trong lis), receivead, so tien): tao transaction")
	fmt.Println("command showlisuser() : show danh sach user ")
	fmt.Println("updatebalancer() : up date so du cac tai khoan cau vi ")
	for {
		fmt.Print("\n> ")
		s := pack.ReadFunc()
		if s == nil {
			fmt.Println("not found command ")
			continue
		}
		n := len(s)

		switch strings.ToLower(strings.TrimSpace(s[0])) {
		case "createuser":
			handleCreateUser(dbusr, s, n)
			break
		case "connectnode":
			handleConnectNode(dbusr, s, n, &NodeAdd)
			break
		case "createtran": // dau vao gom 3 tham so stt dia chi gui , dia chi nhan va so tien gui
			handleCreateTran(dbusr, s, n, NodeAdd)
			break

		case "showlisuser":
			showLisUsr(dbusr)
			break
		case "updatebalancer":
			conn := getConn(NodeAdd)
			if conn == nil {
				fmt.Println("khong ket noi duoc node nao")
				break
			}
			updatebalancer(dbusr, conn)
			break
		default:
			fmt.Println("command error ")
			break
		}

	}

}

// xu ly viec tao transaction va gui transaction toi node
func handleCreateTran(dbusr *leveldb.DB, s []string, n int, NodeAdd []string) {
	if n == 4 {

		stt, err := strconv.Atoi(s[1])
		if err != nil {
			fmt.Println("input err : ", err)
			return
		}

		count, err := strconv.Atoi(s[3])
		if err != nil {
			fmt.Println("input err : ", err)
			return
		}

		if len(s[2]) == 32 {
			ad, err := dbusr.Get(pack.Convert(int64(stt-1)), nil)
			if err != nil {
				fmt.Println("can't read database :  ", err)
				return
			}

			usr, err := dbusr.Get(ad[0:32], nil)

			if err != nil {
				fmt.Println("can't read database :  ", err)
				return
			}
			use := User{}
			err = json.Unmarshal(usr, &use)
			if err != nil {
				fmt.Println("can't Unmarshal json :  ", err)
				return
			}
			receivead := []byte(s[2])
			var receiveAd [32]byte
			copy(receiveAd[:], receivead[0:32])
			tran := CreatTran(use.Add, receiveAd, int64(count), use.PrivateKey)
			conn := getConn(NodeAdd)
			if conn == nil {
				fmt.Println("khong the ket noi den node")
				return
			}
			status := SendTran(tran, conn)
			if status {
				fmt.Println("createtran end")

				return
			} else {
				fmt.Println("createtran error")
				return
			}

		} else {
			fmt.Println("command error")
			return
		}

	} else {
		fmt.Println("command error")
		return
	}
}

// xu ly command connect node
func handleConnectNode(dbusr *leveldb.DB, s []string, n int, NodeAdd *[]string) {
	if n == 2 {
		*NodeAdd = append(*NodeAdd, s[1])
		conn, err := ConnectNode(s[1])
		mss := pack.Message{
			"connnect",
			0,
			nil,
			nil,
			nil,
		}
		mes, err := json.Marshal(mss)
		if err != nil {
			fmt.Println("can't Marshal json Message :  ", err)
			return
		}
		conn.Write(mes)
		conn.Close()
		if err == nil {
			fmt.Println("co the connect toi node")
		} else {
			fmt.Println("khong the connect to node")
		}
		return
	} else {
		fmt.Println("command error")
		return
	}
}

// xu ly command createuser()
func handleCreateUser(dbusr *leveldb.DB, s []string, n int) {
	if n == 2 {
		usr := CreatUser()
		count, err := dbusr.Get([]byte("count"), nil)
		var cin int64
		if err != nil {
			if count == nil {
				count = []byte(pack.Convert(int64(0)))
				cin = 0
			} else {
				fmt.Println("error database")
				return
			}
		} else {
			cin = int64(binary.BigEndian.Uint64(count)) + 1
		}

		b, err := json.Marshal(usr)
		if err != nil {
			fmt.Println("can't Marshal json  :  ", err)
			return
		}

		err = dbusr.Put([]byte("count"), pack.Convert(cin), nil)
		if err != nil {
			fmt.Println("can't write database :  ", err)
			return
		}
		err = dbusr.Put(pack.Convert(cin), usr.Add[0:32], nil)
		if err != nil {
			fmt.Println("can't write database :  ", err)
			return
		}
		err = dbusr.Put(usr.Add[0:32], b, nil)
		if err != nil {
			fmt.Println("can't write database :  ", err)
			return
		}
		fmt.Println("publickey : ", usr.PrivateKey.PublicKey)
		fmt.Println("address user : ", string(usr.Add[0:32]))
		return
	} else {
		fmt.Println("command error")
		return
	}
}

// connect toi mot node trong danh sach cac node da co va  cha lai bien conn
func getConn(s []string) *net.TCPConn {
	n := len(s)
	for i := 0; i < n; i++ {
		conn, err := ConnectNode(s[i])
		if err == nil {
			return conn
		}
	}
	return nil
}

// yeu cau lay so du cua cac tai khoan co trong wallet va cap nhap cac so du do
func updatebalancer(db *leveldb.DB, conn *net.TCPConn) {
	b, err := db.Get([]byte("count"), nil)
	if err != nil {
		fmt.Println("can't read database :  ", err)
		return
	}
	n := int64(binary.BigEndian.Uint64(b))
	acc := []pack.Account{}
	var aci pack.Account
	for i := int64(0); i <= n; i++ {
		ad, err := db.Get(pack.Convert(i), nil)
		if err != nil {
			fmt.Println("can't read database :  ", err)
			return
		}
		copy(aci.Add[:], ad[0:32])
		aci.Balancer = 0
		acc = append(acc, aci)
	}
	mss := pack.Message{
		pack.GetBalancer,
		0,
		nil,
		nil,
		acc,
	}
	mes, err := json.Marshal(mss)
	if err != nil {
		fmt.Println("can't Marshal Message in updatebalancer :  ", err)
		return
	}
	conn.Write(mes)
	bc := make([]byte, 20000)
	count, err := conn.Read(bc)
	if err != nil {
		fmt.Println("can't read response from node :  ", err)
		return
	}
	err = json.Unmarshal(bc[:count], &mss)
	if err != nil {
		fmt.Println("response from node error :  ", err)
		return
	}
	n1 := len(mss.Acc)
	fmt.Println("----------------")
	for i := 0; i < n1; i++ {
		us, err := db.Get(mss.Acc[i].Add[:32], nil)
		if err != nil {
			fmt.Println("can't read database :  ", err)
			return
		}
		usr := User{}
		err = json.Unmarshal(us, &usr)
		if err != nil {
			fmt.Println("can't Unmarshal json from database :  ", err)
			return
		}
		usr.Balancer = mss.Acc[i].Balancer
		usr1, err := json.Marshal(usr)
		if err != nil {
			fmt.Println("can't Marshal json :  ", err)
			return
		}
		db.Put(mss.Acc[i].Add[:32], usr1, nil)

	}
	fmt.Println("update successful")
	conn.Close()

}
