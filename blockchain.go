package blockchain

import (
	"crypto/sha256"
	"fmt"
)

var (
	GenesisAmount float32 = 100
	BlockHeightCounter = 1
)

// Transactions : Represents a single transaction
type Transaction struct {
	Sender string
	Receiver string
	Amount float32
	Fee float32
}


func (tr Transaction) String() string {
	if tr.Sender == "" {
		tr.Sender = "none"
	}
	return "   【 Sender: " + tr.Sender +
				"▕ Receiver: " + tr.Receiver +
				"▕ Amount: " + fmt.Sprintf("%f", tr.Amount) +
				"▕ Fee: " + fmt.Sprintf("%f", tr.Fee) +
				" 】   "
}

// Block : Represents a Block on the Blockchain
type Block struct {
	Transactions  []Transaction
	PrevBlockHash [32]byte
	PrevBlock     *Block
}

// GetBlockHash : Calculates the Sha256 hash of any given Block
func GetBlockHash(block* Block) [32]byte {
	toByte := fmt.Sprintf("%v", block.Transactions) + fmt.Sprintf("%v", block.PrevBlockHash)
	hashVal := sha256.Sum256([]byte(toByte))
	
	return hashVal
}

// InsertBlock : Adds a new Block to the Blockchain
func InsertBlock(transaction []Transaction, chainHead *Block) *Block {
	var hash [32]byte
	
	if chainHead != nil {
		hash = GetBlockHash(chainHead)
	}

	newBlock := Block{Transactions: transaction,
					   PrevBlockHash: hash,
					   PrevBlock: chainHead,
					  }

	// fmt.Println("\nInserting Block ", blockHeightCounter)
	// fmt.Printf("    Block Address: %p\n", &newBlock)
	// fmt.Println(fmt.Sprintf("    Block Hash: %v", GetBlockHash(&newBlock)))
	BlockHeightCounter++
	return &newBlock
}

// ListBlocks : Prints all the Blocks in the Blockchain
func ListBlocks(chainHead *Block) {
	temp := chainHead
	fmt.Println("\n………………… 𝐁𝐥𝐎𝐂𝐊𝐂𝐇𝐀𝐈𝐍 …………………\n")

	for (temp != nil) {
		fmt.Println("Transactions: ")
		for _,txn := range temp.Transactions{
			fmt.Println("              ", txn)
		}
		fmt.Println("\nPrevious Block Hash: ", temp.PrevBlockHash)
		fmt.Printf("Previous Block Address: %p\n", temp.PrevBlock)
		fmt.Printf("Current Block Address: %p\n", temp)
		fmt.Println()
		temp = temp.PrevBlock
	}
}

// GetBalance : Gets the current balance of a node
func GetBalance(chainHead* Block, node string) float32 {
	var amount float32
	
	temp := chainHead
	for (temp != nil) {
		for _, trans := range temp.Transactions {
			if trans.Sender == node {
				amount -= trans.Amount

			} else if (trans.Receiver == node) {
				amount += trans.Amount
			}
		}
		temp = temp.PrevBlock
	}
	return amount
}

func VerifyTxHistory(TxHistory[] Transaction, chain* Block) bool {
	if len(TxHistory) == 0 {
		return true
	}

	temp := chain
	iterator := 0
	for temp != nil {
		for _, trans := range temp.Transactions {
			if trans.Sender == TxHistory[iterator].Sender && trans.Fee == TxHistory[iterator].Fee &&
				trans.Amount == TxHistory[iterator].Amount && trans.Receiver == TxHistory[iterator].Receiver {
				iterator += 1
			}
			if iterator == len(TxHistory) {
				return true
			}
		}
		temp = temp.PrevBlock
	}
	if iterator != len(TxHistory) { return false }
	return false
}

// ValidateTransaction : Validates whether given transaction is valid 
func ValidateTransaction(tr Transaction, chainHead *Block) bool {
	senderBalance := GetBalance(chainHead, tr.Sender)
	return (tr.Sender == "" || senderBalance >= tr.Amount)
}

// DoesStakeExist : Check if given staking txn is part of the Blockchain
func DoesStakeExist(stakeTxn Transaction, chainHead *Block) bool {
	currBlock := chainHead
	for currBlock != nil {
		for _, trans := range currBlock.Transactions {
			if trans == stakeTxn {
				return true
			}
		}
		currBlock = currBlock.PrevBlock
	}
	return false
}

// ValidateTransactions: Validates all the Transactions in the given array
func ValidateTransactions(trans []Transaction, chainHead *Block) bool {
	for _, tr := range trans {
		if ValidateTransaction(tr, chainHead) == false {
			return false
		}
	}
	return true
}
