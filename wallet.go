package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFileName = "wallets.dat"

// a wallet is a public key and a private key
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Wallets is a map from string (the address) to Wallet object
type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
	// make an elliptic curve
	curve := elliptic.P256()

	// generate public/private key
	// I know it just says privatekey but both are inside the struct
	privatekey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println(err)
	}

	// One thing to notice: in elliptic curve based algorithms, public keys
	// are points on a curve. Thus, a public key is
	// a combination of X, Y coordinates.
	// In Bitcoin, these coordinates are concatenated and form a public key.
	publickey := append(privatekey.PublicKey.X.Bytes(), privatekey.PublicKey.Y.Bytes()...)

	return &Wallet{
		PrivateKey: *privatekey,
		PublicKey:  publickey,
	}
}

// loads up a wallets object with all the wallets made so far
// looks for a file, if it doesnt exist, it just returns empty array of *Wallet
func NewWallets() (*Wallets, error) {
	w := Wallets{
		Wallets: make(map[string]*Wallet),
	}

	// check if file exists first, if it doesn't then just
	// load an empty one and return it
	if _, err := os.Stat(walletFileName); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist\n", walletFileName)
		return &w, nil
	}

	// otherwise we have wallets already (in the file). Read the info in
	fileContents, err := os.ReadFile(walletFileName)
	if err != nil {
		log.Panic(err)
	}

	// decode the values in the file using gob
	var wallets Wallets

	// dont rly get why this is necessary, will try to remove later and see
	// if it still works
	gob.Register(elliptic.P256())

	// we need to make a new bytes.Reader here, since NewDecoder expects this
	dec := gob.NewDecoder(bytes.NewReader(fileContents))
	err = dec.Decode(&wallets)
	if err != nil {
		log.Fatal("Decode err:", err)
	}
	return &wallets, nil

}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	sha := sha256.Sum256(pubKey)

	// ripemd160
	riphasher := ripemd160.New()
	riphasher.Write(sha[:])
	publicKeyHash := riphasher.Sum(nil)
	return publicKeyHash
}

// just base 58 DECODES, then removes checksum and version to
// leave us with the public key hash part
func GetPubkeyhashFromAddr(address string) []byte {
	decoded := base58.Decode(address)
	return decoded[1 : len(decoded)-4]
}

// generates a bitcoin address using a Wallet's public key
func (w *Wallet) generateAddress() []byte {
	publicKeyHash := HashPubKey(w.PublicKey)

	checksum1 := sha256.Sum256(publicKeyHash)
	checksum2 := sha256.Sum256(checksum1[:])

	output := append([]byte{version}, publicKeyHash...)
	output = append(output, checksum2[:]...)

	encoded := []byte(base58.Encode(output))

	return encoded
}

// creates a new wallet to add to the wallets object
func (w *Wallets) createWallet() string {
	wallet := NewWallet()
	address := string(wallet.generateAddress())
	w.Wallets[address] = wallet

	return address
}

// find a wallet (private/pub key pair) given a human readable address
// internally this just calculates the hash of the public key
// for each wallet and sees if its equal to the address
func (w *Wallets) findWallet(address string) Wallet {
	return *w.Wallets[address]
}

// this serializes the Wallets object and writes to the file
func (w *Wallets) saveToFile() {
	var output bytes.Buffer
	gob.Register(elliptic.P256())
	enc := gob.NewEncoder(&output)
	err := enc.Encode(w)
	if err != nil {
		log.Fatal("Encode err:", err)
	}
	err = os.WriteFile(walletFileName, output.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
