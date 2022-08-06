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
	fmt.Println("  listaddresses - list all the addresses on this network")
	fmt.Println("  createwallet - Generates a public/private keypair, returns your address")
	fmt.Println("  clear - Clears all the files (blockchain.db) and (wallets.dat)")
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
	createWallet := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddresses := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	clear := flag.NewFlagSet("clear", flag.ExitOnError)

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
	case "createwallet":
		err := createWallet.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddresses.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "clear":
		err := clear.Parse(os.Args[2:])
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

	if createWallet.Parsed() {
		cli.createWallet()
	}

	if listAddresses.Parsed() {
		cli.listAddresses()
	}

	if clear.Parsed() {
		cli.clear()
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

	// create UTXO Set
	utxoset := UTXOSet{
		Blockchain: blockchain,
	}
	// and initialize it
	utxoset.Reindex()
}

// the essence of sending is two parts:
// 1. deducting "amount" from from's balance
// 2. adding "amount" to to's balance

// The first part is done by populating the TXInput array on the new transaction
// to POINT to the TXOutput that we are consuming
// (using TXInput's Txid and OutputIdx fields)
// so that later when FindUnspentTransactions runs during something like getBalance,
// it will recognize the previously unspent outputs as spent, and the balance
// will effectively be "deducted" from from's account, because spent
// outputs are skipped when counting total balance. That's the whole idea.

// the second part is done by making available a bunch of new UTXOs,
// whose 'PublicKeyHash' field is the corresponding hash of the person
// to receive the money. And again, when getBalance runs again, these
// new outputs are not tied to any input and hence are added to the balance
// of the owner
func (cli *CLI) send(from, to string, amount int) {

	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	blockchain := InitBlockchain(from)
	defer blockchain.DB.Close()

	// create transaction
	transaction := NewGeneralTransaction(from, to, amount, blockchain)

	// This is the "miners reward" in our network, to keep it simple, let's say
	// the person who sends the transaction will get the reward
	// for mining, although in a real implementation this obviously
	// wouldn't be the case
	minerReward := NewCoinbaseTX(from, "")

	// create and add new block to chain (this does the mining)
	block := blockchain.AddBlock([]*Transaction{transaction, minerReward})

	// update UTXO set
	UTXOSet := UTXOSet{
		Blockchain: blockchain,
	}
	UTXOSet.Update(block)

	fmt.Println("Successfully sent", amount, "from", from, "to", to)
}

func (cli *CLI) getBalance(address string) {

	ret := 0
	blockchain := InitBlockchain(address)
	defer blockchain.DB.Close()

	// create UTXO Set
	UTXOSet := UTXOSet{
		Blockchain: blockchain,
	}
	unspentTransactionOutputs := UTXOSet.FindUTXO(GetPubkeyhashFromAddr(address))

	for _, output := range unspentTransactionOutputs {
		ret += output.Value
	}
	fmt.Printf("The address %s has %d balance currently\n", address, ret)
}

func (cli *CLI) createWallet() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addr := wallets.createWallet()
	fmt.Printf("Made a wallet, your address is %s", addr)

	wallets.saveToFile()
}

// this just goes thru all the Wallet objects in Wallets
// and creates addresses from the public keys
func (cli *CLI) listAddresses() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	for address, _ := range wallets.Wallets {
		cli.getBalance(address)
	}
}

func (cli *CLI) clear() {
	e := os.Remove("blockchain.db")
	if e != nil {
		fmt.Println(e)
	}
	e2 := os.Remove("wallets.dat")
	if e2 != nil {
		fmt.Println(e2)
	}
}
