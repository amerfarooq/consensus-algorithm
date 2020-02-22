package main

import (
	bc "assignment02IBC/ibc_project/blockchain"
	"assignment02IBC/ibc_project/network"
	"assignment02IBC/ibc_project/network/protocol"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)



var (
	port 		 string = ":3005"
	chainHead    *bc.Block
	matureStakes []network.Stake
	agingStakes  []network.Stake
	minAge    	 int = 2
	mutex 		 sync.Mutex
	latestStake  network.Stake
	isWaiting	 bool = false
)

func receiveStake(nodeMsg *network.Message) {
	recvStake := nodeMsg.Content.(network.Stake)
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
	txn := bc.Transaction{
		Sender:   recvStake.Sender.Name,
		Receiver: "director",
		Amount:   recvStake.Amount,
		Fee:      0,
	}
	chainHead = bc.InsertBlock([]bc.Transaction{txn}, chainHead)
	fmt.Println(">> Blockchain updated ")
	bc.ListBlocks(chainHead)

	mutex.Lock()
	agingStakes = append(agingStakes, recvStake)
	mutex.Unlock()

	mintBlock()
}

func ageStakes() {
	for true {
		time.Sleep(5 * time.Second)
		mutex.Lock()
		for _, stake := range matureStakes {
			stake.Age++
		}
		var temp [] network.Stake
		for _, stake := range agingStakes {
			stake.Age++
			if stake.Age >= minAge {
				matureStakes = append(matureStakes, stake)
			} else {
				temp = append(temp, stake)
			}
		}
		agingStakes = temp
		fmt.Println("\n>> Aging stakes: ", agingStakes)
		fmt.Println(">> Mature stakes: ", matureStakes)

		mutex.Unlock()
		if len(matureStakes) > 0 {
			mintBlock()
		}
	}
}

func selectMinter() network.Stake {
	fmt.Println("\n………………… Staking Process Begins …………………\n")

	var worths []int
	var max int

	for _, stake := range matureStakes {
		worths =  append(worths, int(stake.Amount) + len(stake.TxnHistory) + stake.Age)
		max += int(stake.Amount) + len(stake.TxnHistory) + stake.Age
	}
	fmt.Println(">> Mature stakes", matureStakes)
	fmt.Println(">> Maximum value", max)

	var ranges []int
	start := 0
	for _, worth := range worths {
		ranges = append(ranges,  start + worth)
		start = start + worth
	}
	fmt.Println(">> Worth of mature stakes", ranges)

	validatorIndex := rand.Intn(max + 1)
	stakeIndex := 0
	var validator network.Stake
	for _, worth := range ranges {
		if worth >= validatorIndex {
			validator = matureStakes[stakeIndex]
		}
		stakeIndex++
	}
	fmt.Println(">> Random Number: ", validatorIndex)
	fmt.Println(">> Selected Validator: ", validator)
	fmt.Println("\n………………… Staking Process Ends …………………\n")

	return validator
}

func mintBlock() {
	if len(matureStakes) == 0 {
		fmt.Println(">> No stake has matured yet")
		return
	} else if isWaiting {
		fmt.Println(">> Still waiting for reply from minter")
		return
	}
	fmt.Println(">> Mature stakes exist")
	fmt.Println("   >> ", matureStakes)

	stakeOfMinter := selectMinter()
	network.SendMessage(protocol.Mine_Block, nil, stakeOfMinter.Sender)
	latestStake = stakeOfMinter
	isWaiting = true

	var temp []network.Stake
	for _, stake := range matureStakes {
		if  stake.Sender == stakeOfMinter.Sender &&
			stake.Amount == stakeOfMinter.Amount &&
			stake.TimeOfCreation == stakeOfMinter.TimeOfCreation {
			continue
		} else {
			temp = append(temp, stake)
		}
	}
	matureStakes = temp
	stakeOfMinter = network.Stake{}
}

func listen() {
	listener, err := net.Listen("tcp4", port)
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		nodeMsg := network.ReceiveMessage(conn)

		fmt.Println("\n             ˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍ")
		fmt.Println("≡≡≡≡≡≡≡≡≡≡≡≡≡| New Message |≡≡≡≡≡≡≡≡≡≡≡≡≡")
		fmt.Println("             ˉˉˉˉˉˉˉˉˉˉˉˉˉˉˉ")

		switch nodeMsg.Protocol {
			case protocol.Receive_Stake:
				receiveStake(&nodeMsg)

			case protocol.Receive_Block:
				receiveBlock(&nodeMsg)

			case protocol.Receive_Genesis_Block:
				receiveGenesisBlock(&nodeMsg)
		}
	}
}

func isBlockValid(recvBlock bc.Block) bool {
	fmt.Println(">> Validating received Block")

	transactions := recvBlock.Transactions
	var reward float32

	for _, txn := range transactions[:len(transactions) - 1] {
		if !bc.ValidateTransaction(txn, chainHead) {
			fmt.Println(">> Invalid transaction")
			fmt.Println(txn)
			return false
		}
		reward += txn.Fee
	}
	fmt.Println(">> Calculated Reward:", reward)
	if reward != transactions[len(transactions)-1].Amount {
		fmt.Println(">> Invalid reward amount!")
		return false
	}
	return true
}

func receiveBlock(nodeMsg *network.Message) {
	fmt.Println(">> Received block")
	recvBlock := nodeMsg.Content.(bc.Block)

	if isBlockValid(recvBlock) {
		fmt.Println(">> Block is valid")
		fmt.Println(">> Blockchain updated")
		chainHead = &recvBlock
		bc.ListBlocks(chainHead)
		go releaseStake(latestStake)
		network.SendMessage(protocol.Flood_Block, recvBlock, latestStake.Sender)
	}
	isWaiting = false
	latestStake = network.Stake{}
}

func releaseStake(stake network.Stake) {
	for len(matureStakes) == 0 && len(agingStakes) == 0 {
		fmt.Println(">> Stake cannot be released yet!")
		time.Sleep(10 * time.Second)
	}
	fmt.Println("\n>> Releasing stake!")
	fmt.Println(">> Stake: ", stake)

	releaseStakeTxn := bc.Transaction{
		Sender:   "director",
		Receiver:  stake.Sender.Name,
		Amount:   stake.Amount,
		Fee:      0,
	}
	chainHead = bc.InsertBlock([]bc.Transaction{releaseStakeTxn}, chainHead)

	for _, node := range agingStakes {
		network.SendMessage(protocol.Receive_Release_Stake, *chainHead, node.Sender)
	}
	for _, node := range matureStakes {
		network.SendMessage(protocol.Receive_Release_Stake, *chainHead, node.Sender)
	}
	fmt.Println(">> Blockchain updated")
	bc.ListBlocks(chainHead)
}

func receiveGenesisBlock(nodeMsg *network.Message) {
	recvBlock := nodeMsg.Content.(bc.Block)

	fmt.Println(">> Receiving the genesis Block")
	chainHead = &recvBlock
	fmt.Println(">> Blockchain updated")
	bc.ListBlocks(chainHead)
}

func main(){
	go listen()
	go ageStakes()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
