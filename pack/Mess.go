package pack

const (
	SendBlock       = "SEND_BLOCK"
	GetHeighBlock   = "GET_HEIGH"
	DistributeBlock = "DISTRIBUTE_BLOCK"
	GetBlock        = "RECEIVE_BLOCK" // thong diep nhan block bat dau tu block heigh trong mess
	SendHeighBlock  = "SEND_HEIGH"
	CreateTran      = "CREATE_TRAN"
	GetBalancer     = "GET_BALANCER"
)

type Message struct {
	Cmd   string
	Heigh int64
	Data  []Block
	Tran  []Transaction
	Acc   []Account
}

type Account struct {
	Add      [32]byte
	Balancer int64
}
