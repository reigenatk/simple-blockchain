package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
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
}

func (b *Block) SetHash() {
	// first send timestamp int64 to string, then string to []byte
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))

	// basically this just adds the byte slices together into one big []byte
	// kinda ugly way to do it but it works
	headers := bytes.Join([][]byte{timestamp, b.Data, b.PrevBlockHash}, []byte{})

	// take SHA256 hash
	hash := sha256.Sum256(headers)

	// convert [32]byte to []byte
	b.Hash = hash[:]
}

// a function to create a new block given some data that the block should store
// and the previous block hash
func NewBlock(data string, prevBlockHash []byte) *Block {
	ret := Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
	}
	ret.SetHash()
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
