package main

import (
	"log"
	"strconv"
)

// we want 24 zero bits or more to accept a hash as OK
const targetBits = 24

func main() {
	blockchain := InitBlockchain()
	blockchain.AddBlock("some info 1")
	blockchain.AddBlock("some info 2")

	for _, block := range blockchain.blocks {
		powChecker := NewProofOfWork(block)
		isValid := powChecker.Validate()
		log.Printf("Is block with data %s valid? %s", block.Data, strconv.FormatBool(isValid))
	}
}
