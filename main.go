package main

// we want 24 zero bits or more to accept a hash as OK
const targetBits = 24

func main() {
	blockchain := initBlockchain()
	blockchain.AddBlock("some info 1")
	blockchain.AddBlock("some info 2")

}
