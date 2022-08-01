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
	addBlock := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChain := flag.NewFlagSet("printchain", flag.ExitOnError)

	// add data arg to addBlock command
	addBlockData := addBlock.String("data", "", "Data for new block")

	// call Parse depending on what the subcommand is?
	switch os.Args[1] {
	case "addblock":
		err := addBlock.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChain.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	}

	// if it was to addblock
	if addBlock.Parsed() {
		// if they didnt pass any data to add, exit out
		if *addBlockData == "" {
			addBlock.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	// if it was to print chain
	if printChain.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
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
		fmt.Printf("Block with hash %x, Data: \"%s\" Prev Hash: %x, PoW: %s\n\n", block.Hash, block.Data, block.PrevBlockHash, strconv.FormatBool(isValid))

		// terminate when the previous block hash is empty
		// meaning we are at the genesis block
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
