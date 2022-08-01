package main

type TXInput struct {
	// a list of all the TXOutputs it references
	Txid []byte
	// stores an index of an output in the transaction
	Vout int
	// Data to be used with an output's ScriptPubKey
	ScriptSig string
}
