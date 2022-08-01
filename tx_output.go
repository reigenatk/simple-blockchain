package main

type TXOutput struct {
	// the amount of money
	Value int
	// the puzzle used to lock the transaction
	// usually a random list of strings
	ScriptPubKey string
}

func (txo *TXOutput) CanBeUnlockedWith(unlockData string) bool {
	return txo.ScriptPubKey == unlockData
}
