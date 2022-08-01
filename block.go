package main

import (
	"bytes"
	"encoding/gob"
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
	log.Printf("nonce is %d, hash is %x", nonce, hash)

	return &ret
}

// the first block on the chain
func GenesisBlock() *Block {
	return NewBlock("Genesis Data", []byte{})
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
