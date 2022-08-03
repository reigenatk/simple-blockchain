package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const blocksBucket string = "blocks"
const genesisBlockData string = "Genesis Block"

// a blockchain can be entirely defined by
// 1. the hash of the latest block
// 2. the connection to the database (which we only want one instance of)
type Blockchain struct {
	LatestHash []byte
	DB         *bolt.DB
}

// an iterator for looping thru the blocks in our blockchain in order
// since bolt stores keys by byte-order which isn't the right order
// blockchainIterator will go from latest block
// to oldest (top to bottom so to speak)
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// add a new block to the blockchain, takes in a list of transactions
// to tie to the block we're adding. This also saves it to the DB automatically
func (bc *Blockchain) AddBlock(transactions []*Transaction) {

	var LatestHash []byte

	// try to find what the latest hash was
	// this will be "previousHash" field for this new block we're making
	bc.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		LatestHash = bucket.Get([]byte("l"))
		return nil
	})

	// make the new block
	b := NewBlock(transactions, LatestHash)

	// write the hash of this new block into DB as latest hash
	bc.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		bucket.Put(b.Hash, b.Serialize())
		bucket.Put([]byte("l"), b.Hash)

		// also update the blockchain struct accordingly
		bc.LatestHash = b.Hash
		return nil
	})
}

// this function is weird in the sense that we're not actually
// initializing a blockchain on each call. Instead, we're looking
// in the DB to see if it exists. Only if it doesnt exist do we
// initialize, so basically it only gets created once. The rest of the time
// we're just getting the latest hash from the DB
// and putting using that hash to return a Blockchain object for use
// RULES:
// 32-byte block-hash -> Block structure (serialized)
// 'l' -> the hash of the last block in a chain (l for latest)
func InitBlockchain(address string) *Blockchain {

	// hash of the tip of the blockchain (latest block)
	var tip []byte

	// first open database file
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	// start read write transaction in Bolt
	err = db.Update(func(tx *bolt.Tx) error {
		// try to get the "Block" bucket
		blockbucket := tx.Bucket([]byte(blocksBucket))

		// if this bucket doesnt exist and is nil
		// then it means we haven't initialized the blockchain
		// (aka it has no blocks and its probably our first run of this program)
		// so make the genesis block and write it
		// into the blockchain, also put it as last hash
		if blockbucket == nil {
			fmt.Println("No blockchain detected, creating genesis block...")

			// create coinbase transaction to put on genesis block
			// the unlock key for this transaction is the address.
			newTransaction := NewCoinbaseTX(address, genesisBlockData)
			firstBlock := GenesisBlock(newTransaction)

			// make a new block
			b, _ := tx.CreateBucket([]byte(blocksBucket))

			err = b.Put(firstBlock.Hash, firstBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = b.Put([]byte("l"), firstBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = firstBlock.Hash
		} else {
			// otherwise we have a blockchain already
			// get the topmost block
			tip = blockbucket.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	// make the blockchain struct
	blockchain := Blockchain{
		LatestHash: tip,
		DB:         db,
	}

	return &blockchain
}

// function to make a blockchain iterator
// sort of "captures" a blockchain in a certain state so to speak
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: bc.LatestHash,
		db:          bc.DB,
	}
}

// returns the block pointed to by the blockchainIterator
// this is found by comparing the currentHash field
// it also has the side effect of moving the blockchainIterator
// to point to the next Block
func (bci *BlockchainIterator) Next() *Block {
	var block *Block

	err := bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		dbBlock := bucket.Get([]byte(bci.currentHash))
		block = Deserialize(dbBlock)
		return nil
	})

	if err != nil {
		log.Println(err.Error())
		log.Panic(err)
	}

	bci.currentHash = block.PrevBlockHash
	return block
}

// looks for unspent transactions by checking every block in the blockchain
// Use TXInputs to add to the map, TXOutputs to check the map
// if it isn't in the map, then its unspent. Check if the address unlocks
// the TXOutput, and if so, we have an unspent transaction
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction

	// map from string to int slice
	// or transaction ID to index of spent outputs
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		// check this block's transactions
		for _, tx := range block.Transactions {
			// get the string format of a transaction ID
			// we will use this as the key to the map spentTXOs
			txID := hex.EncodeToString(tx.ID)

			// use a label so we can return here
		Outputs:
			// go over each output transaction in this transaction
			for outIdx, out := range tx.Vout {
				// if this transaction has any spent outputs the
				// results []int will not be nil
				// fmt.Printf("spentTXOs for %s is %v\n", txID, spentTXOs[txID])
				if spentTXOs[txID] != nil {

					for _, spentOut := range spentTXOs[txID] {
						// compare the index of the current vOut
						// to that of the values in the array, if
						// any are matching then this specific output
						// transaction is matched, and therefore spent
						// by an input. So we're done with this one

						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				// if any of the output transactions are not
				// matched, then put the whole TRANSACTION object into unspentTXs array
				// which signifies that in this transaction there is still some money
				// to spend, that belongs to address.
				// also couldnt we add the same transaction object MULTIPLE times,
				// one for each unmatched txoutput...?
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			// mark the outputs that the inputs of the current transaction
			// connect to as "spent". We do this by accesing the Vout field of
			// each TXInput in tx.Vin, which is the index of the Vout, and
			// adding to the list there each time the input can be unlocked
			// with the address. Finally we're putting this list under
			// the hash of the output's transaction's ID
			if !tx.isCoinbase() {

				for _, in := range tx.Vin {

					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						fmt.Printf("Adding %d to spentTXOs\n", in.Vout)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		// if at genesis block, then we're done
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}

// this is just like find unspent transactions
// but since that only returns Transaction objects
// we want something more specific, the actual TXOutputs
func (bc *Blockchain) findUnspentTXOs(address string) []TXOutput {
	var ret []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)
	// fmt.Printf("Unspent transactions for %s is %v\n", address, unspentTransactions)

	// loop through each transaction object, which we know must have
	// a few unspent outputs
	for _, transaction := range unspentTransactions {
		for _, txoutput := range transaction.Vout {
			if txoutput.CanBeUnlockedWith(address) {
				ret = append(ret, txoutput)
			}
		}
	}
	return ret
}

// finds all UTXOs under address' name, and adds up the balances
// until its greater than the amount we need
// also use a map to store which TXOutput objects we needed to use
func (bc *Blockchain) findSpendableOutputs(address string, amount int) (int, map[string][]int) {
	ret := make(map[string][]int)

	// first find all transactions which have unspent money owned by address
	unspentTransactions := bc.FindUnspentTransactions(address)

	balance := 0

Work:
	for _, tx := range unspentTransactions {
		txID := hex.EncodeToString(tx.ID)
		for outputidx, output := range tx.Vout {
			// check if this output belongs to our address
			if output.CanBeUnlockedWith(address) && balance < amount {
				balance += output.Value

				// mark this output transaction as used
				ret[txID] = append(ret[txID], outputidx)

				// this is just to save some time
				if balance >= amount {
					break Work
				}
			}
		}
	}

	return balance, ret
}
