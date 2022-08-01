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

// add a new block to the blockchain
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

		// if this is nil then it means we haven't initialized the blockchain
		// yet (aka it has no blocks), so make the genesis block and write it
		// into the blockchain, also say its the last hash
		if blockbucket == nil {
			fmt.Println("No blockchain detected, creating genesis block...")

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
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction

	// map from string to int slice
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		// check this block's transactions
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			// use a label so we can return here
		Outputs:
			// for each output transaction
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.isCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}
