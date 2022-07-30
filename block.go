package main

import (
	"bytes"
	"log"
	"time"
)

// In Bitcoin specification, Timestamp, PrevBlockHash, and Hash are
// block headers, which form a separate data structure, and
// transactions (Data in our case) is a separate data structure.
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// a function to create a new block given some data that the block should store
// and the previous block hash
func NewBlock(data string, prevBlockHash []byte) *Block {
	ret := Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
	}
	// first ask proof of work to find the right nonce and hash
	// for this block
	pow := NewProofOfWork(&ret)
	nonce, hash := pow.Run()
	ret.Hash = hash[:]
	ret.Nonce = nonce
	log.Printf("nonce is %d, hash is %v", nonce, hash)

	return &ret
}

// the first block on the chain
func GenesisBlock() *Block {
	ret := Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte("genesis"),
		PrevBlockHash: []byte{},
	}
	return &ret
}

// format the block header into []byte
// including the nonce (which is the miner's guess)
func (b *Block) prepareHashBytes(nonce int) []byte {
	timestamp := intToBuffer(b.Timestamp)
	target := intToBuffer(targetBits)
	nonceBytes := intToBuffer(int64(nonce))
	headers := bytes.Join([][]byte{timestamp, b.Data, b.PrevBlockHash, target, nonceBytes}, []byte{})
	return headers
}
