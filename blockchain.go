package main

// a blockchain is just a slice of Block pointers!
type Blockchain struct {
	blocks []*Block
}

// add a new block to the blockchain
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]

	// newblock is going to return *Block, the block to add
	b := NewBlock(data, prevBlock.Hash)

	// add this block to our slice of blocks, aka the blockchain
	// append returns the updated slice, so we have to set again
	bc.blocks = append(bc.blocks, b)
}

func initBlockchain() *Blockchain {
	// get the genesis block
	firstBlock := GenesisBlock()

	// put the genesis block as the first block manually
	blockchain := Blockchain{
		blocks: []*Block{firstBlock},
	}

	return &blockchain
}
