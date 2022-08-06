package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const UTXOSetbucket = "UTXOSet"

// the idea behind the UTXOSet is to just avoid having to scan
// the entire blockchain for transactions. It's like a cache of sorts
// but it needs to be kept consistent with the real blockchain
type UTXOSet struct {
	Blockchain *Blockchain
}

// this function kinda acts as a "refresh" for the UTXO set.
// this will go through the entire blockchain via findAllUnspentTXOs
// so we want to call this sparingly, only when necessary
func (utxos *UTXOSet) Reindex() {
	db := utxos.Blockchain.DB

	// delete this bucket (erases all previously held data about the UTXO set)
	_ = db.Update(func(tx *bolt.Tx) error {
		_ = tx.DeleteBucket([]byte(UTXOSetbucket))
		tx.CreateBucket([]byte(UTXOSetbucket))
		return nil
	})

	UTXO := utxos.Blockchain.findAllUnspentTXOs()

	_ = db.Update(func(tx *bolt.Tx) error {
		// try to get the "Block" bucket
		bucket := tx.Bucket([]byte(UTXOSetbucket))

		// put each unspent TXoutput into Bolt
		for txID, TXoutput := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put(key, TXoutput.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

// gives you balance of an address, as well as which transaction outputs make up
// this balance
func (utxos *UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := utxos.Blockchain.DB

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(UTXOSetbucket))

		// this is standard way to loop thru bucket
		// look in Bolt docs
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			// first deserialize []byte to TXOutputs
			outputs := DeserializeOutputs(v)

			// check if the output is unlockable via this pubkeyHash
			for idx, output := range outputs.Outputs {
				if output.IsLockedWithKey(pubkeyHash) {
					unspentOutputs[txID] = append(unspentOutputs[txID], idx)
					accumulated += output.Value
				}
			}
		}
		return nil
	})
	return accumulated, unspentOutputs
}

// just like FindSpendableOutputs except we don't return amount or map,
// we just return a list of the actual TXOutput objects.
func (utxos *UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := utxos.Blockchain.DB
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(UTXOSetbucket))

		// this is standard way to loop thru bucket
		// look in Bolt docs
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			// first deserialize []byte to TXOutputs
			outputs := DeserializeOutputs(v)

			// check if the output is unlockable via this pubkeyHash
			for _, output := range outputs.Outputs {
				if output.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, output)
				}
			}
		}
		return nil
	})
	return UTXOs
}

// inform the UTXO Set about a new block that has appeared on the chain
// call this right after we add a block to the blockchain
func (utxos *UTXOSet) Update(block *Block) {
	db := utxos.Blockchain.DB

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXOSetbucket))

		// loop over each transaction in this newly added block
		for _, tx := range block.Transactions {
			// we don't add coinbase transactions to UTXO set
			if tx.isCoinbase() {
				continue
			}

			// for each input, check which outputs it references. Removes
			// those referenced outputs from the UTXO set, since they are no longer
			// unspent
			for _, vin := range tx.Vin {
				outputToRemoveIdx := vin.OutputIdx
				newTxOutputs := TXOutputs{}
				curTxOutputs := b.Get(vin.Txid)
				txOutputs := DeserializeOutputs(curTxOutputs)

				fmt.Printf("Removing output %d from transaction %s\n", outputToRemoveIdx, vin.Txid)
				// remove this specific workflow by adding everything but this
				// newly removed output. If golang had a remove element from slice
				// by value, this would be equivalent to that. But idk if it does
				for idx, output := range txOutputs.Outputs {
					if idx != outputToRemoveIdx {
						newTxOutputs.Outputs = append(newTxOutputs.Outputs, output)
					}
				}

				// if there are no more outputs left for this transaction, don't
				// bother updating the DB since there's nothing in the 'value'
				// part of key/value
				if len(newTxOutputs.Outputs) == 0 {
					b.Delete(vin.Txid)
				} else {
					// delete old value and write new one into DB
					b.Put(vin.Txid, newTxOutputs.Serialize())
				}
			}

			// ok, we've removed stale outputs. Now to add new outputs from
			// this block! All outputs are guaranteed unspent since we just made
			// the block before getting here
			newTxOutputs := TXOutputs{}
			newTxOutputs.Outputs = append(newTxOutputs.Outputs, tx.Vout...)
			b.Put(tx.ID, newTxOutputs.Serialize())
		}
		return nil
	})
}
