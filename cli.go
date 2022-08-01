package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI responsible for processing command line arguments
type CLI struct {
	bc *Blockchain
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("addblock [data]: Adds a block to the chain with data")
	fmt.Println("printchain: Prints out the entire blockchain")
}

func (cli *CLI) validateArgLength() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgLength()

	// define two possible commands
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChain := flag.NewFlagSet("printchain", flag.ExitOnError)
	newBlockchain := flag.NewFlagSet("newblockchain", flag.ExitOnError)
	getBalance := flag.NewFlagSet("getbalance", flag.ExitOnError)

	// extra args
	getBalanceAddress := getBalance.String("address", "", "address to get balance from")
	createBlockchainAddress := newBlockchain.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	// call Parse depending on what the subcommand is?
	switch os.Args[1] {
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChain.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "newblockchain":
		err := printChain.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if getBalance.Parsed() {
		if *getBalanceAddress == "" {
			getBalance.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	// if it was to print chain
	if printChain.Parsed() {
		cli.printChain()
	}

	if newBlockchain.Parsed() {
		cli.InitBlockchain(*createBlockchainAddress)
	}
}

func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock()
	fmt.Println("Block successfully added")
}

// prints out each block in the chain
func (cli *CLI) printChain() {
	curIterator := cli.bc.Iterator()

	for {
		// this returns the current block that iterator points to
		// despite the name being Next, it just moves the iterator next one
		block := curIterator.Next()

		// validate if the block is valid once again
		powChecker := NewProofOfWork(block)
		isValid := powChecker.Validate()

		// print all our findings
		fmt.Printf("Block with hash %x, Prev Hash: %x, PoW: %s\n\n", block.Hash, block.PrevBlockHash, strconv.FormatBool(isValid))

		// terminate when the previous block hash is empty
		// meaning we are at the genesis block
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) InitBlockchain(address string) {
	blockchain := InitBlockchain(address)
	defer blockchain.DB.Close()

}

func (cli *CLI) send(from, to string, amount int) {
	blockchain := InitBlockchain(from)
	defer blockchain.DB.Close()
}

func (cli *CLI) getBalance(address string) {

}
