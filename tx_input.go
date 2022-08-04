package main

import "bytes"

// The first two fields has to deal with
// the TXOutput that this TXInput references
// a single TXInput can reference only one TXOutput
// from a previous transaction
type TXInput struct {
	// this stores the ID of the transaction that
	// the corresponding TXOutput belongs to
	// The ID should be encoded to hex when
	// used as a string
	Txid []byte
	// stores the index of an TXOutput in the transaction
	OutputIdx int
	Signature []byte
	PublicKey []byte
}

// just hashes the public key on the input
// and checks if its equal to the argument
func (txi *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(txi.PublicKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
