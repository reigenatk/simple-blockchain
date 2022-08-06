package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TXOutput struct {
	// the amount of money
	Value         int
	PublicKeyHash []byte
}

func (txo *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(txo.PublicKeyHash, pubKeyHash) == 0
}

func (txo *TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(txo)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// we store txoutputs (PLURAL) aka multiple outputs using gob encoder
type TXOutputs struct {
	Outputs []TXOutput
}

func DeserializeOutputs(outputbytes []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(outputbytes))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
