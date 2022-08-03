package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/itchyny/base58-go"
)

type TXOutput struct {
	// the amount of money
	Value         int
	PublicKeyHash []byte
}

// sets the public key hash on a TXOutput equal to the hash of an address
// this would be used for instance, when we send coins to someone.
func (txo *TXOutput) Lock(address []byte) {
	encoding := base58.BitcoinEncoding
	decoded, err := encoding.Decode(address)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// we want to skip version and checksum
	txo.PublicKeyHash = decoded[1 : len(decoded)-4]
}

func (txo *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(txo.PublicKeyHash, pubKeyHash) == 0
}
