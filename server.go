package main

// import (
// 	"fmt"
// 	"net"
// )

// // When a new node is run, it gets several nodes from a DNS seed,
// // and sends them version message,
// // which in our implementation will look like this:
// type Version struct {
// 	// version of the client we are running
// 	// we only have one version so this doesnt really do anything
// 	Version int
// 	// Length of this node's blockchain
// 	BestHeight int
// 	// The address of the sender
// 	AddrFrom string
// }

// var nodeAddress string
// var knownNodes = []string{"localhost:3000"}

// func StartServer(nodeID, minerAddress string) {
// 	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
// 	miningAddress := minerAddress
// 	ln, _ := net.Listen(protocol, nodeAddress)
// 	defer ln.Close()

// 	bc := InitBlockchain(nodeID)

// 	// if current node is not the central one, it must send version message
// 	// to the central node to find out if its blockchain is outdated.
// 	if nodeAddress != knownNodes[0] {
// 		sendVersion(knownNodes[0], bc)
// 	}

// 	for {
// 		conn, err := ln.Accept()
// 		go handleConnection(conn, bc)
// 	}
// }

// func sendVersion(addr string, bc *Blockchain) {
// 	bestHeight := bc.GetBestHeight()
// 	payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})

// 	request := append(commandToBytes("version"), payload...)

// 	sendData(addr, request)
// }
