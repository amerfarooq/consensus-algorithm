package main

import (
	"assignment02IBC/ibc_project/network"
	"assignment02IBC/ibc_project/network/protocol"
	bc "assignment02IBC/ibc_project/blockchain"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	port string = ":3000"
	totalNodes int
	connectedNodes int = 0
	peers []network.Peer
	isGenesisBlockMined bool = false
)


func handleNewConnectingNode(nodeMsg* network.Message) {
	fmt.Println(">> Handling a new connection")
	connectedNodes++
	peers = append(peers, nodeMsg.Content.(network.Peer))

	if (!isGenesisBlockMined) {
		fmt.Println(">> Asking", nodeMsg.Content.(network.Peer).Name, " to mine the Genesis Block")
		delegateGenesisBlockCreation(nodeMsg.Content.(network.Peer))
		isGenesisBlockMined = true
	}
	if connectedNodes == totalNodes {
		fmt.Println(">> All", totalNodes, "nodes have connected")
		sendAddresses()
	} 
}

func delegateGenesisBlockCreation(node network.Peer) {
	trans := bc.Transaction{"", node.Name, bc.GenesisAmount, 0}
	network.SendMessage(protocol.Mine_Genesis_Block, trans, node)
}


func handleConnection(conn net.Conn) {
	nodeMsg := network.ReceiveMessage(conn)
	fmt.Println(">> Connection message content: ", nodeMsg.Content)
	fmt.Println(">> Connection message protocol: ", nodeMsg.Protocol)

	switch nodeMsg.Protocol {
		case protocol.New_Connection:
			handleNewConnectingNode(&nodeMsg)
	}
	conn.Close()
}

func getNodePeers(index int) []network.Peer {
	currentPeer := peers[index]
	totalPeers := rand.Intn(len(peers))

	if (totalPeers == 0) {
		totalPeers++
	}
	var nodePeers []network.Peer
	peersAdded := 0

	for peersAdded != totalPeers {
		possiblePeer := peers[rand.Intn(len(peers))]

		if possiblePeer != currentPeer {
			nodePeers = append(nodePeers, possiblePeer)
			peersAdded++
		}
	}
	return nodePeers
}


func sendAddresses() {
	fmt.Println(">> Sending addresses to all nodes")

	for index, peer := range peers {
		peers := getNodePeers(index)
		network.SendMessage(protocol.Receive_Addresses, peers, peer)
	}
}

func init() {
	rand.Seed(time.Now().Unix())
}


func main() {
	arguments := os.Args
	
	totalNodes, _ = strconv.Atoi(arguments[1])
	listener, _ := net.Listen("tcp4", port)

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("\n             ˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍˍ")
		fmt.Println(  "≡≡≡≡≡≡≡≡≡≡≡≡≡| New Connection |≡≡≡≡≡≡≡≡≡≡≡≡≡")
		fmt.Println("             ˉˉˉˉˉˉˉˉˉˉˉˉˉˉˉˉˉˉ")

		go handleConnection(conn)
	}
}