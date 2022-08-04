package main

import (
	"bytes"

	"github.com/btcsuite/btcutil/base58"
)

type TXOutput struct {
	// the amount of money
	Value         int
	PublicKeyHash []byte
}

// sets the public key hash on a TXOutput equal to the hash of an address
// this would be used for instance, when we send coins to someone.
func (txo *TXOutput) Lock(address []byte) {

	decoded := base58.Decode(string(address))

	// we want to skip version and checksum
	txo.PublicKeyHash = decoded[1 : len(decoded)-4]
}

func (txo *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(txo.PublicKeyHash, pubKeyHash) == 0
}
