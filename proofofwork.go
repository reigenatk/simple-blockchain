package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

// the largest value of nonce we can try, 2^63
const maxNonce int64 = math.MaxInt64

// the target field helps us decide whether or not a block's hash
// is valid. If the hash <= target, then its valid, else, try again
// this is because we're looking for a certain # of leading zeros
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// create a new Proof of Work for a specific Block
func NewProofOfWork(b *Block) *ProofOfWork {
	// use the math/big package to deal with large numbers
	// this sets target = 1 << (256-targetBits)
	// we are doing 256 because we'll use SHA256
	// which has 256 bit output
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	return &ProofOfWork{
		b,
		target,
	}
}

// core loop
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	// mine for the right nonce
	for int64(nonce) < maxNonce {
		// get the bytes for the hash
		hashbytes := pow.block.prepareHashBytes(nonce)

		// perform the hash
		hash = sha256.Sum256(hashbytes)

		// send hash to big.Int form so we can compare
		hashInt.SetBytes(hash[:])

		// if hash as integer is less than target
		// we've found a working nonce
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			// else keep trying (mining)
			nonce++
		}
	}
	return nonce, hash[:]
}

// check if said block's nonce + block header evaluates to
// something smaller than the target. This more or less does same
// calculation as Run except this time we are checking the nonce
// instead of searching for a valid one
func (pow *ProofOfWork) Validate() bool {
	hashbytes := pow.block.prepareHashBytes(pow.block.Nonce)

	hash := sha256.Sum256(hashbytes)

	var hashInt big.Int

	hashInt.SetBytes(hash[:])

	if hashInt.Cmp(pow.target) == -1 {
		return true
	} else {
		return false
	}
}
