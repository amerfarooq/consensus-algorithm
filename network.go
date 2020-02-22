package network

import (
	bc "assignment02IBC/ibc_project/blockchain"
	"assignment02IBC/ibc_project/network/protocol"
	"encoding/gob"
	"log"
	"net"
	"time"
)

type Message struct {
    Protocol protocol.Protocol 
    Content  interface{}
}

type Peer struct {
    Name string
    Address string
}

type Stake struct {
	Sender         Peer
	Amount         float32
	TxnHistory     []bc.Transaction
	Age            int
	TimeOfCreation time.Time
}

func SendMessage(prot protocol.Protocol, content interface{}, receiver Peer) {
	conn, err := net.Dial("tcp4", "localhost:" + receiver.Address)

	if err != nil {
		log.Fatal(err)
	}
	message := Message{
		Protocol: prot,
		Content: content,
	}
	gob.Register(bc.Block{})
	gob.Register(Peer{})
    gob.Register([]Peer{})
    gob.Register(bc.Transaction{})
	gob.Register(Stake{})

	enc := gob.NewEncoder(conn)
	err = enc.Encode(message)

	if err != nil {
		log.Fatal(err)
	}
	conn.Close()	
}

func ReceiveMessage(conn net.Conn) Message {
	var nodeMsg Message
	
	gob.Register(bc.Block{})
	gob.Register(Peer{})
	gob.Register([]Peer{})
	gob.Register(Stake{})
	
	err := gob.NewDecoder(conn).Decode(&nodeMsg)

	if err != nil {
		log.Fatal(err)
	}
	return nodeMsg
}