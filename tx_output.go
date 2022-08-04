package main

import (
	"bytes"
)

type TXOutput struct {
	// the amount of money
	Value         int
	PublicKeyHash []byte
}

func (txo *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(txo.PublicKeyHash, pubKeyHash) == 0
}
