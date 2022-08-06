package main

import (
	"bytes"
	"crypto/ecdsa"
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
// to set equal to the "Transactions" field of
// the block we're adding. This also saves it to the DB automatically

func (bc *Blockchain) AddBlock(transactions []*Transaction) *Block {

	var LatestHash []byte

	// before we add it to the chain though, we must VERIFY the digital signature
	// on all TXInputs for each Transaction.
	for _, tx := range transactions {
		isVerified := bc.verifyTransaction(tx)
		if !isVerified {
			log.Panic("Block failed digital signature verification, exiting!")
		}
	}

	// try to find what the latest hash was, we need it since
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
	return b
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
	db, err := bolt.Open("blockchain.db", 0600, nil)
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

// kinda like FindUnspentTransactions but instead there's no argument
// and we don't check if the output belongs to a certain person
func (bc *Blockchain) findAllUnspentTXOs() map[string]TXOutputs {
	unspentTXs := make(map[string]TXOutputs)

	// map from string to int slice
	// or transaction ID to index of spent outputs
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
			txoutputs := TXOutputs{}
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				txoutputs.Outputs = append(txoutputs.Outputs, out)
				unspentTXs[txID] = txoutputs
			}

			if !tx.isCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					// fmt.Printf("Adding %d to spentTXOs\n", in.OutputIdx)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.OutputIdx)
				}
			}

		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}

// gets a certain transaction given a transaction ID
// does this simply by using the Iterator and going through all the blocks
func (bc *Blockchain) findTransaction(id []byte) (Transaction, error) {
	it := bc.Iterator()
	for {
		block := it.Next()
		for _, transaction := range block.Transactions {
			if bytes.Compare(id, transaction.ID) == 0 {
				return *transaction, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, fmt.Errorf("No transaction of this ID was found!")
}

// this just creates the map needed to call transaction.Sign()
// by repeatedly calling findTransaction, and then once that map is
// ready we just call Sign with what we just made
// takes in a private key and ID of transaction to sign, which makes sense
func (bc *Blockchain) signTransaction(tx *Transaction, privkey ecdsa.PrivateKey) {

	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		transaction, _ := bc.findTransaction(vin.Txid)
		transactionReferenced := hex.EncodeToString(transaction.ID)
		prevTXs[transactionReferenced] = transaction
	}
	tx.Sign(privkey, prevTXs)
}

// this verifies a digital signature on a transaction
func (bc *Blockchain) verifyTransaction(tx *Transaction) bool {

	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		transaction, err := bc.findTransaction(vin.Txid)
		if err != nil {
			log.Print(err)
		}
		transactionReferenced := hex.EncodeToString(transaction.ID)
		prevTXs[transactionReferenced] = transaction
	}
	return tx.Verify(prevTXs)
}

// func (bc *Blockchain) GetBestHeight() int {

// }
