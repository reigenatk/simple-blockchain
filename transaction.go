package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

// how much we reward a miner for a block
// aka the coinbase transaction
const subsidy = 10

// a transaction consists of an ID and a lists of inputs + outputs
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// coinbase transaction is a specific type of transaction
// used to reward miners. It is not a normal transaction
// in the sense that it accesses no previous outputs
// and also stores a subsidy (miner reward) as the value in its output
func NewCoinbaseTX(to, data string) *Transaction {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	toWallet := wallets.findWallet(to)

	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	// here we set the publicKey on the TXInput to be
	// equal to some random string since it wasn't signed by anyone
	// nor does it have to be verified (we will avoid calling
	// verify on it using isCoinbase() function)
	txin := TXInput{
		Txid:      []byte{},
		OutputIdx: -1,
		PublicKey: []byte(data),
		Signature: nil,
	}
	txout := TXOutput{
		Value:         subsidy,
		PublicKeyHash: HashPubKey(toWallet.PublicKey),
	}
	tx := &Transaction{
		ID:   nil,
		Vin:  []TXInput{txin},
		Vout: []TXOutput{txout},
	}
	tx.setID()
	return tx
}

// makes a new transaction object to
// transfer x money from account a to b
// the inputs "spend" the money from the sender
// and the output is a new unspent transaction with "amount" money
// unlockable only by the receiver's address, "to"
func NewGeneralTransaction(from, to string, amount int, blockchain *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// get a list of all wallets
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	// find the wallet that has the "from" address
	// we do this because we need to use the public/private key
	fromWallet := wallets.findWallet(from)

	// need this since we wanna try to unlock unspent transactions
	// code word for verifying the digital signatures
	pubKeyHash := HashPubKey(fromWallet.PublicKey)

	amountOwned, outputTransactions := blockchain.findSpendableOutputs(pubKeyHash, amount)

	// check if enough money
	if amountOwned < amount {
		log.Panic("Not enough balance!")
	}

	// take all the output transactions used to get this balance
	// and make it the new inputs
	for txid, listIdxes := range outputTransactions {
		txidbytes, _ := hex.DecodeString(txid)
		for _, idx := range listIdxes {
			input := TXInput{
				Txid:      txidbytes,
				OutputIdx: idx,
				PublicKey: fromWallet.PublicKey,
				Signature: nil, // signing this transaction will populate this fie
			}
			inputs = append(inputs, input)
		}
	}

	// make ScriptPubKey "to" so that the money belongs to "to" now

	output := TXOutput{
		Value:         amount,
		PublicKeyHash: GetPubkeyhashFromAddr(to),
	}
	outputs = append(outputs, output)

	// if we weren't exact (which is likely, say we needed to send 50 but we had
	// only +20 and +40) then we refund the extra 10 back to the sender, "from"
	if amountOwned > amount {
		output := TXOutput{
			Value:         amountOwned - amount,
			PublicKeyHash: GetPubkeyhashFromAddr(from),
		}
		outputs = append(outputs, output)
	}

	tx := &Transaction{
		ID:   nil,
		Vin:  inputs,
		Vout: outputs,
	}

	// this populates the ID field
	tx.setID()

	// sign the whole transaction, aka imprint our privateKey on it
	// this will auto populate the TXInput's "Signature" field
	blockchain.signTransaction(tx, fromWallet.PrivateKey)
	return tx
}

// sets the transaction ID on a transaction to the sha256 hash of the
// entire transaction
func (tx *Transaction) setID() {
	// remove ID field first
	txcopy := *tx
	txcopy.ID = []byte{}

	var output bytes.Buffer
	enc := gob.NewEncoder(&output)
	err := enc.Encode(txcopy) // use the copy which has no ID
	if err != nil {
		log.Fatal("Transaction encode err:", err)
	}
	hash := sha256.Sum256(output.Bytes())
	tx.ID = hash[:]
}

// we can tell that a transaction is a coinbase type if
// the vin array has length 1 and the OutputIdx is -1, and Txid of that transaction is
// of length 0. Just as we set in NewCoinbaseTX
func (tx *Transaction) isCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].OutputIdx == -1
}

// prevTXs is just a map from transaction IDs (hex strings) to Transaction object
// This func goes to each input in this transaction, finds the corresponding
// Transaction which has the output, grabs the hash from that TXOutput
// and gets an ID (by hashing the encoded Transaction object),
// then we run ecdsa.Sign with that ID AS THE DATA,
// and finally set the Signature field to whatever that value is. Phew.
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.isCoinbase() {
		return
	}
	txtrim := tx.TrimmedCopy()
	for idx, vin := range txtrim.Vin {
		// find the Transaction referenced by each TXInput
		prevtxID := hex.EncodeToString(vin.Txid)

		// gets actual object
		prevTX := prevTXs[prevtxID]

		// Sets the PublicKey field on the TXInput to the PublicKeyHash
		// on the output of some previous Transaction object
		txtrim.Vin[idx].PublicKey = prevTX.Vout[vin.OutputIdx].PublicKeyHash

		// set the ID of the Transaction to the hash
		// this internally calls gob, so that's why we trimmed earlier
		txtrim.setID()

		// set it back to nil
		txtrim.Vin[idx].PublicKey = nil

		// perform the digital signature
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txtrim.ID)

		if err != nil {
			log.Panic(err)
		}

		// get signature and store in original Transaction Input!
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[idx].Signature = signature
	}
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
// literally the same as before except no Signature and PubKey fields on Inputs
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.OutputIdx, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PublicKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// this verifies each TXInput on a transaction object by using the
// signature and the public key, and calling ecdsa.Verify()
// this function probably should be called after Sign(), otherwise
// it makes no sense
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {

	// no sense in verifying coinbase transactions
	if tx.isCoinbase() {
		return true
	}

	// check to see if all transactions have IDs on them, they should by now
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txtrim := tx.TrimmedCopy()
	curve := elliptic.P256()

	// verify EACH input
	for idx, vin := range tx.Vin {
		// same procedure as in Sign
		prevtxID := hex.EncodeToString(vin.Txid)
		prevTX := prevTXs[prevtxID]
		txtrim.Vin[idx].Signature = nil
		txtrim.Vin[idx].PublicKey = prevTX.Vout[vin.OutputIdx].PublicKeyHash
		txtrim.setID()
		txtrim.Vin[idx].PublicKey = nil

		sigLen := len(vin.Signature)
		r := big.Int{}
		s := big.Int{}
		// r and s are split equally
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		// get the public key (its on the TXInput object)
		pubKeyLen := len(vin.PublicKey)
		x := big.Int{}
		y := big.Int{}
		// r and s are split equally
		x.SetBytes(vin.PublicKey[:(pubKeyLen / 2)])
		y.SetBytes(vin.PublicKey[(pubKeyLen / 2):])

		pubKey := ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}

		// do the verify
		isVerified := ecdsa.Verify(&pubKey, txtrim.ID, &r, &s)
		if !isVerified {
			return false
		}
	}
	return true
}
