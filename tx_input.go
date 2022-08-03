package main

import "bytes"

type TXInput struct {
	// a list of all the TXOutputs it references
	// a single TXInput can reference only one TXOutput
	// from a previous transaction
	// this stores the ID of that transaction
	Txid []byte
	// stores the index of an TXOutput in the transaction
	OutputIdx int
	Signature []byte
	PublicKey []byte
}

// hashes the public key on the input and checks if its equal to the argument
func (txi *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(txi.PublicKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
