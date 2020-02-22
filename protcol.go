package protocol

type Protocol int

const (
	New_Connection Protocol = iota
	Receive_Blockchain 
	Receive_Addresses
	Receive_Transaction
	Mine_Block
	Validate_Block
	Receive_Release_Stake
	Mine_Genesis_Block
	Receive_Genesis_Block
	Receive_Stake
	Receive_Block
	Flood_Block
)

func (p Protocol) String() string {
	return [...]string{"New Connection", "Receive Blockchain", "Receive Addresses", "Receive Transactions",
		               "Mine New Block", "Validate Block", "Mine Genesis Block", "Receive Genesis Block",
		               	"Receive Stake", "Validate Stake", "Receive Block"}[p]
}