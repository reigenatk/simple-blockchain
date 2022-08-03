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
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  newblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
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
	newBlockchainAddress := newBlockchain.String("address", "", "The address to send genesis block reward to")
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
		err := newBlockchain.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalance.Parse(os.Args[2:])
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
		cli.InitBlockchain(*newBlockchainAddress)
	}
}

// prints out each block in the chain
func (cli *CLI) printChain() {
	if cli.bc == nil {
		cli.bc = InitBlockchain("default")
	}
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
	// if blockchain already exists this does nothing basically
	blockchain := InitBlockchain(address)
	defer blockchain.DB.Close()

}

// the essence of sending is two parts:
// 1. deducting "amount" from from's balance
// 2. adding "amount" to to's balance
// the first part is done by making available a new UTXOs, unspent transaction output
// with value equal to "amount" and ScriptPubKey equal to "to"
// the second part is done by populating the TXInput array on the new transaction
// so that when FindUnspentTransactions runs during something like getBalance,
// it will recognize the previously unspent outputs as spent, and the balance
// will effectively be "deducted" from from's account, because spent
// outputs are skipped when counting total balance. That's the whole idea.
func (cli *CLI) send(from, to string, amount int) {
	blockchain := InitBlockchain(from)
	defer blockchain.DB.Close()

	// create transaction
	transaction := NewGeneralTransaction(from, to, amount, blockchain)
	// create and add new block to chain
	blockchain.AddBlock([]*Transaction{transaction})
	fmt.Println("Successfully sent", amount, "from", from, "to", to)
}

func (cli *CLI) getBalance(address string) {

	ret := 0
	blockchain := InitBlockchain(address)
	defer blockchain.DB.Close()
	unspentTransactionOutputs := blockchain.findUnspentTXOs(address)

	for _, output := range unspentTransactionOutputs {
		ret += output.Value
	}
	fmt.Printf("The address %s has %d balance currently", address, ret)
}
