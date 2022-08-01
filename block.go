package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

// In Bitcoin specification, Timestamp, PrevBlockHash, and Hash are
// block headers, which form a separate data structure, and
// transactions (Data in our case) is a separate data structure.
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// a function to create a new block given some data that the block should store
// and the previous block hash
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	ret := Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
	}
	// first ask proof of work to find the right nonce and hash
	// for this block
	pow := NewProofOfWork(&ret)
	nonce, hash := pow.Run()
	ret.Hash = hash[:]
	ret.Nonce = nonce
	log.Printf("nonce is %d, hash is %x", nonce, hash)

	return &ret
}

// the first block on the chain
func GenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// a function to serialize the Block struct to a []byte so we can
// store it inside the DB. We use encoding/gob package to do the encoding
// for us, its very efficient.
func (b *Block) Serialize() []byte {
	var output bytes.Buffer
	enc := gob.NewEncoder(&output)
	err := enc.Encode(b)
	if err != nil {
		log.Fatal("Encode err:", err)
	}
	return output.Bytes()
}

// opposite of Serialize, has to take a Block from the database and
// put it back into our Block struct
func Deserialize(b []byte) *Block {
	var block Block

	// we need to make a new bytes.Reader here, since NewDecoder expects this
	dec := gob.NewDecoder(bytes.NewReader(b))
	err := dec.Decode(&block)
	if err != nil {
		log.Fatal("Decode err:", err)
	}
	return &block
}

// takes the transactions field off a Block object and
// transforms it into a []byte so we can
// put it with the other block header info and
// start guessing nonces to mine the block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	// add each transaction's ID only
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	// join all the IDs into a big byte slice
	// and take the hash of it
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}
