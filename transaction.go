package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

// how much we reward a miner for a block
// aka the coinbase transaction
const subsidy = 10

// a transaction consists of an ID and a lists of inputs + outputs
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// coinbase transaction is a specific type of transaction
// used to reward miners. It has one input only
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{
		Value:        subsidy,
		ScriptPubKey: to,
	}
	tx := &Transaction{
		ID:   nil,
		Vin:  []TXInput{txin},
		Vout: []TXOutput{txout},
	}
	tx.setID()
	return tx
}

// sets the transaction ID on a transaction to the sha256 hash of the
// entire transaction
func (tx *Transaction) setID() {
	var output bytes.Buffer
	enc := gob.NewEncoder(&output)
	err := enc.Encode(tx)
	if err != nil {
		log.Fatal("Transaction encode err:", err)
	}
	hash := sha256.Sum256(output.Bytes())
	tx.ID = hash[:]
}
