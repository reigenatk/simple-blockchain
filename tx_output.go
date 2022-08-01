package main

type TXOutput struct {
	// the amount of money
	Value int
	// the puzzle used to lock the transaction
	// usually a random list of strings
	ScriptPubKey string
}
