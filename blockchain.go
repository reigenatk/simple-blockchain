package main

import (
	"log"

	"github.com/boltdb/bolt"
)

const blocksBucket string = "blocks"

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
func (bc *Blockchain) AddBlock(data string) {

	var LatestHash []byte

	// try to find what the latest hash was
	// this will be "previousHash" field for this new block we're making
	bc.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		LatestHash = bucket.Get([]byte("l"))
		return nil
	})

	// make the new block
	b := NewBlock(data, LatestHash)

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
func InitBlockchain() *Blockchain {

	// hash of the tip of the blockchain (latest block)
	var tip []byte

	// first open database file
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// start read write transaction in Bolt
	err = db.Update(func(tx *bolt.Tx) error {
		// try to get the "Block" bucket
		blockbucket := tx.Bucket([]byte(blocksBucket))

		// if this is nil then it means we haven't initialized the blockchain
		// yet (aka it has no blocks), so make the genesis block and write it
		// into the blockchain, also say its the last hash
		if blockbucket == nil {
			firstBlock := GenesisBlock()

			// make a new block
			b, _ := tx.CreateBucket([]byte(blocksBucket))

			err = b.Put(firstBlock.Hash, firstBlock.Serialize())
			err = b.Put([]byte("l"), firstBlock.Hash)
			tip = firstBlock.Hash
		} else {
			// otherwise we have a blockchain already
			// get the topmost block
			b := tx.Bucket([]byte(blocksBucket))
			tip = b.Get([]byte("l"))
		}

		return nil
	})
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

// given a blockchainIterator, go to the next one
// in order to go next one we merely need to CHANGE the current hash
// to the value of the previous hash,
// which is stored in the previous block! So
// first obtain the block and deserialize it
// then look at the previousHash field
func (bci *BlockchainIterator) Next() *BlockchainIterator {
	bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		dbBlock := bucket.Get([]byte(bci.currentHash))
		block := Deserialize(dbBlock)
		bci.currentHash = block.PrevBlockHash
		return nil
	})
	return bci
}
