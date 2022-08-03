## Writing simple blockchain in Go

Following [this](https://jeiwan.net/posts/building-blockchain-in-go-part-1/) I will attempt to write a simple blockchain. I know the basics of crypto (for example all the buzzwords) but decided that to *really* understand it, I should probably write up an implementation.

### Theory

Bitcoin uses **Hashcash** as a hashing protocol. It works by taking the block header (which consists of Timestamp, previous hash and data) and then adding a nonce to it, merging all that into a byte array, and taking the hash. If the hash meets a requirement as put forth by the current difficulty value, it will pass. Otherwise increment the nonce and try again. 

In this implementation we use SHA256 as our hashing algorithm (which lots of cryptos use today as well), which takes in any input and outputs 256 bytes. A hashing algorithm has a few important properties, the main one is that given the output of the hash, it is nearly impossible to figure out the input. The formal term is **one-way**. This makes it computationally impossible for anyone to change the blockchain, hence giving its security.

The **nonce** is just like the "guess", and the goal of the miners is to find a **working nonce**. This might take a lot of guesses depending on the difficulty, which is why miners have to have good hardware specs. The guessing process is literally just starting with a value of 0 and going up to infinity. This process of finding a suitable nonce is also known as **mining**. Every time a new block is added to the blockchain, it must be mined.

The **difficulty** is another way of saying that we set a certain target. For instance:
> The first three bytes of the hash must be 0

Then when we want to **verify** whether a given nonce (which generates a specific hash) passes the target, we can check whether that hash is less than `1<<(256 - 24)`, 256 because of SHA256 and -24 because then if any of the first 3 bytes are nonzero, it will automatically be larger and we will instantly know this nonce isn't good enough. Obviously, this 3 value is up for our choosing, and that is where the concept of *difficulty* comes in- a larger value would indicate a higher difficulty since the hashes have to be even smaller to pass. 

The concept of having to mine each block is called **Proof of Work**, basically we can have confidence in our records because each block was computationally verified. This obviously is very useful for official records like financial balances and transactions, hence why blockchain is so closely tied to cryptocurrency.

Blockchains also have **persistance**, meaning a database of some sort. We use [BoltDB](https://github.com/boltdb/bolt) in this implementation, a key-value store written in Go. Values are stored in **buckets**, and we will use two kinds of key -> value pairs (a simplified version of real Bitcoin implementation). They are:

> 32-byte block-hash -> Block structure (serialized)
> 'l' -> the hash of the last block in a chain

TX = "transaction"
UTXO = "Unspent Transaction Outputs", basically any outputs that have not been referenced by any inputs. We care about these because that is our currency!
Lock/Unlock- transaction outputs and inputs are locked by address values. This is how we distinguish who owns which money.

### Code

Blockchain has blocks (which you can access using Iterator), each block has a list of transactions, and each transaction has a list of inputs/outputs (TXInput, TXOutput). Inputs on the transaction reference previous transactions' outputs, but only one input to correspont to one output and vice versa.

Delete the .db file to create a new blockchain.

Libraries we use
- `encoding/gob` for easy serialization/deserialization
- `boltDB` for persistance
- `flag` for user input
