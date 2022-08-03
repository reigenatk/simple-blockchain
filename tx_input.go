package main

type TXInput struct {
	// a list of all the TXOutputs it references
	// a single TXInput can reference only one TXOutput
	// from a previous transaction
	Txid []byte
	// stores an index of an output in the transaction
	Vout int
	// Data to be used with an output's ScriptPubKey
	ScriptSig string
}

func (txi *TXInput) CanUnlockOutputWith(unlockData string) bool {
	return txi.ScriptSig == unlockData
}
