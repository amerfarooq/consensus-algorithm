package main

import (
	bc "assignment02IBC/ibc_project/blockchain"
	"assignment02IBC/ibc_project/network"
	"assignment02IBC/ibc_project/network/protocol"
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	port string
	networkInitiator network.Peer = network.Peer{"Network Initiator", "3000"}
	director network.Peer = network.Peer{"Director", "3005"}
	peers []network.Peer
	name string
	chainHead *bc.Block = nil
	minedGenesisBlock bool = false
	memPool []bc.Transaction = nil
	txnHistory []bc.Transaction					// Verified transactions that are part of the Blockchain
	seenStakes []network.Stake
)


func listen() {
	listener, err := net.Listen("tcp4", ":" + port)
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()
	
	for {		
		conn, _ := listener.Accept()
		nodeMsg := network.ReceiveMessage(conn)

		fmt.Println("\n             ˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍ")
		fmt.Println(  "≡≡≡≡≡≡≡≡≡≡≡≡≡| New Message |≡≡≡≡≡≡≡≡≡≡≡≡≡")
		fmt.Println(  "             ˉˉˉˉˉˉˉˉˉˉˉˉˉˉˉ")

		switch nodeMsg.Protocol {
			case protocol.Receive_Addresses:
				receivePeers(&nodeMsg)
				if (minedGenesisBlock) {
					propogateGenesisBlockToPeers()
					network.SendMessage(protocol.Receive_Genesis_Block, *chainHead, director)
				}
				
			case protocol.Mine_Block:
				mineBlock(&nodeMsg)

			case protocol.Validate_Block:
				validateBlock(&nodeMsg)

			case protocol.Receive_Stake:
				receiveStake(&nodeMsg)

			case protocol.Mine_Genesis_Block:
				mineGenesisBlock(&nodeMsg)

			case protocol.Receive_Genesis_Block:
				receiveGenesisBlock(&nodeMsg)

			case protocol.Receive_Transaction:
				receiveTransaction(&nodeMsg)

			case protocol.Flood_Block:
				fmt.Println(">> Flooding Block")
				propogateChainToPeers()

			case protocol.Receive_Release_Stake:
				receiveReleaseStake(&nodeMsg)
		}
	}
}

func receiveReleaseStake(nodeMsg *network.Message) {
	recvBlock := nodeMsg.Content.(bc.Block)
	fmt.Println(">> Received stake release block")

	if bc.GetBlockHash(&recvBlock) == bc.GetBlockHash(chainHead) {
		fmt.Println(">> Block already included in blockchain")
	} else {
		fmt.Println(">> Blockchain updated")
		chainHead = &recvBlock
		bc.ListBlocks(chainHead)
		propogateReleaseStake()
	}
}

func propogateReleaseStake() {
	for _, peer := range peers {
		network.SendMessage(protocol.Receive_Release_Stake, *chainHead, peer)
		fmt.Println(">> Sending", peer.Name, "release stake Block")
	}
}


func receiveStake(nodeMsg *network.Message) {
	recvStake := nodeMsg.Content.(network.Stake)
	stakingTxn := bc.Transaction{
		Sender:   recvStake.Sender.Name,
		Receiver: "director",
		Amount:   recvStake.Amount,
		Fee:      0,
	}
	if bc.DoesStakeExist(stakingTxn, chainHead){
		fmt.Println(">> Already received this stake!")
		return
	}
	fmt.Println(">> Received a new stake: ", recvStake)

	stakerBalance := bc.GetBalance(chainHead, recvStake.Sender.Name)
	fmt.Println(">> Staker balance: ", stakerBalance)

	if stakerBalance <= recvStake.Amount {
		fmt.Println(">> Staker does not have the staked amount!")
		return
	} else if !bc.VerifyTxHistory(recvStake.TxnHistory, chainHead) {
		fmt.Println(">> Staker did not perform the claimed transactions!")
		return
	}
	fmt.Println(">> Stake is valid")
	chainHead = bc.InsertBlock([]bc.Transaction{stakingTxn}, chainHead)
	fmt.Println(">> Blockchain updated")
	bc.ListBlocks(chainHead)
	propogateStakeToPeers(recvStake)
}

func stake(stakeAmt float32) {
	myStake := network.Stake{
		Sender:         network.Peer{name, port},
		Amount:         stakeAmt,
		TxnHistory:     txnHistory,
		Age:            0,
		TimeOfCreation: time.Now(),
	}
	network.SendMessage(protocol.Receive_Stake, myStake, director)

	txn := bc.Transaction{
		Sender:   name,
		Receiver: "director",
		Amount:   stakeAmt,
		Fee:      0,
	}
	chainHead = bc.InsertBlock([]bc.Transaction{txn}, chainHead)
	fmt.Println(" >> Blockchain updated")
	bc.ListBlocks(chainHead)
	propogateStakeToPeers(myStake)
}

func receivePeers(nodeMsg* network.Message) {
	peers = nodeMsg.Content.([]network.Peer)
	fmt.Println(">> Received peers from network manager: ", peers)
}

func receiveTransaction(nodeMsg* network.Message) {
	recvTrans := nodeMsg.Content.(bc.Transaction)
	fmt.Println(">> Received transaction of", recvTrans.Sender)
	fmt.Println(">> Transaction: ", recvTrans)

	for _, tx := range memPool {
        if tx == recvTrans {
			fmt.Println(">> Transaction already part of Mempool!")
            return
        }
	}
	memPool = append(memPool, recvTrans)
	printMempool()
	propogateTransactionToPeers(recvTrans)
}

func validateBlock(nodeMsg* network.Message) {
	recvBlock := nodeMsg.Content.(bc.Block)
	fmt.Println(">> Validating received Block")

	if bc.GetBlockHash(&recvBlock) == bc.GetBlockHash(chainHead) {
		fmt.Println(">> Block already validated")
	} else {
		transactions := recvBlock.Transactions
		var reward float32
		fmt.Println(">> TXNS: ", transactions)

		for _, txn := range transactions[:len(transactions) - 1] {
			fmt.Println(">> Validating transaction")
			fmt.Println(">> txn", txn)
			if !bc.ValidateTransaction(txn, chainHead) {
				fmt.Println(">> Invalid transaction")
				fmt.Println(txn)
				return
			}
			if txn.Sender == name {
				txnHistory = append(txnHistory, txn)
			}
			amendMemPool(txn)
			reward += txn.Fee
		}
		fmt.Println(">> Calculated reward:", reward)
		fmt.Println(">> Claimed reward:", transactions[len(transactions)-1].Amount)

		if reward != transactions[len(transactions)-1].Amount {
			fmt.Println(">> Invalid reward amount!")
			return
		}
		fmt.Println(">> Blockchain updated")
		chainHead = &recvBlock
		bc.ListBlocks(chainHead)

		propogateChainToPeers()
	}
}

func amendMemPool(validTxn  bc.Transaction) {
	var temp []bc.Transaction
	updated := false

	for _, txn := range memPool {
		if txn != validTxn {
			temp = append(temp, txn)
		} else {
			updated = true
		}
	}
	memPool = temp

	if updated {
		printMempool()
	}
}

func mineGenesisBlock(nodeMsg* network.Message) {
	fmt.Println(">> Mining the Genesis block")
	trans := nodeMsg.Content.(bc.Transaction)
	fmt.Println(">> Received Genesis Transactions:", trans)
	
	chainHead = bc.InsertBlock([]bc.Transaction{trans}, chainHead)
	fmt.Println(">> Blockchain updated")
	bc.ListBlocks(chainHead)
	minedGenesisBlock = true
}

func receiveGenesisBlock(nodeMsg* network.Message) {
	recvBlock := nodeMsg.Content.(bc.Block)

	if chainHead != nil {
		fmt.Println(">> Already received the Genesis Block")
		return
	}
	fmt.Println(">> Receiving the Genesis Block")
	chainHead = &recvBlock
	fmt.Println(">> Blockchain updated")
	bc.ListBlocks(chainHead)
	propogateGenesisBlockToPeers()		
}

func mineBlock(nodeMsg* network.Message) {
	fmt.Println(">> Mining new block")

	if len(memPool) == 0 {
		fmt.Println(">> " + name + "'s Mempool is empty. New block will not be mined!")
		return
	}
	var reward float32
	var validTxns []bc.Transaction

	for _, txn := range memPool {
		if !bc.ValidateTransaction(txn, chainHead) {
			fmt.Println(">> Transaction is invalid", txn)
			continue
		}
		fmt.Println(">> Transaction is valid", txn)
		validTxns = append(validTxns, txn)
		reward += txn.Fee
	}
	coinBaseTrans := bc.Transaction{Receiver: name, Amount: reward}
	validTxns = append(validTxns, coinBaseTrans)
	chainHead = bc.InsertBlock(validTxns, chainHead)
	fmt.Println(">> Blockchain updated")
	bc.ListBlocks(chainHead)
	network.SendMessage(protocol.Receive_Block, *chainHead, director)
}

func propogateChainToPeers() {
	for _, peer := range peers {
		network.SendMessage(protocol.Validate_Block, *chainHead, peer)
		fmt.Println(">> Asking", peer.Name, "to validate mined Block")
	}
}

func propogateStakeToPeers(stake network.Stake) {
	for _, peer := range peers {
		network.SendMessage(protocol.Receive_Stake, stake, peer)
		fmt.Println(">> Sending", peer.Name, "stake")
	}
}

func propogateTransactionToPeers(tran bc.Transaction) {
	for _, peer := range peers {
		network.SendMessage(protocol.Receive_Transaction, tran, peer)
		fmt.Println(">> Sending", peer.Name, "the transaction")
	}
}

func propogateGenesisBlockToPeers() {
	for _, peer := range peers {
		network.SendMessage(protocol.Receive_Genesis_Block, *chainHead, peer)
		fmt.Println(">> Sending", peer.Name, "the Genesis Block")
	}
}

func getUserInput() {
	for {
		fmt.Println("\n")
		reader := bufio.NewReader(os.Stdin)

		inputType, _ := reader.ReadString('\n')
		inputType = strings.TrimSuffix(inputType, "\n")

		if inputType == "TX" {
			receiver, _ := reader.ReadString('\n')
			receiver = strings.TrimSuffix(receiver, "\n")

			var amount float32
			fmt.Scanf("%f", &amount)
			if amount <= 0 {
				fmt.Println(">> Transaction rejected. Amount has to be greater than 0!")
				continue
			}

			var fee float32
			fmt.Scanf("%f", &fee)
			if amount <= 0 {
				fmt.Println(">> Transaction rejected. Fee has to be greater than 0!")
				continue
			}
			trans := bc.Transaction{Sender: name, Receiver: receiver, Amount: amount, Fee: fee}
			memPool = append(memPool, trans)
			printMempool()
			propogateTransactionToPeers(trans)

		} else if inputType == "STK" {
			var amount float32
			fmt.Scanf("%f", &amount)

			if amount <= 0 {
				fmt.Println(">> Stake amount has to be greater than 0!")
				continue
			}
			stake(amount)
		}
	}
}

func printMempool() {
	fmt.Println(">> Mempool updated")
	fmt.Println("\n………………… MEMPOOL …………………\n")
	for _, txn := range memPool {
		fmt.Println(txn)
	}
	fmt.Println()
}

func main() {
	arguments := os.Args
	port = arguments[1]
	name = arguments[2]

	network.SendMessage(protocol.New_Connection, 
					    network.Peer{Name: name, Address: port},
						networkInitiator,
	)

	go getUserInput()
	go listen()
	
	wg := &sync.WaitGroup{}
    wg.Add(1)
    wg.Wait()
}
    