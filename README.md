## Writing simple blockchain in Go

Following [this](https://jeiwan.net/posts/building-blockchain-in-go-part-1/) I will attempt to write a simple blockchain. I know the basics of crypto (for example all the buzzwords) but decided that to *really* understand it, I should probably write up an implementation. After all, that's the best way to really understand something. Also, this implementation will try to follow Bitcoin as much as possible.

### Concepts

Bitcoin uses **Hashcash** as a hashing protocol. It works by taking the block header (which consists of Timestamp, previous hash and data) and then adding a nonce to it, merging all that into a byte array, and taking the hash. If the hash meets a requirement as put forth by the current difficulty value, it will pass. Otherwise increment the nonce and try again. 

In this implementation we use SHA256 as our hashing algorithm (which lots of cryptos use today as well), which takes in any input and outputs 256 bytes. A hashing algorithm has a few important properties, the main one is that given the output of the hash, it is nearly impossible to figure out the input. The formal term is **one-way**. This makes it computationally impossible for anyone to change the blockchain, hence giving its security.

### Proof of Work

The **nonce** is just like the "guess", and the goal of the miners is to find a **working nonce**. This might take a lot of guesses depending on the difficulty, which is why miners have to have good hardware specs. The guessing process is literally just starting with a value of 0 and going up to infinity. This process of finding a suitable nonce is also known as **mining**. Every time a new block is added to the blockchain, it must be mined.

The **difficulty** is another way of saying that we set a certain target. For instance:
> The first five bytes of the hash must be 0

![image](https://user-images.githubusercontent.com/69275171/182676783-a08665c2-d86e-472d-97e4-42aabbf1ce1b.png)

Then when we want to **verify** whether a given nonce (which generates a specific hash) passes the target, we can check whether that hash is less than `1<<(256 - 40)`, 256 because of SHA256 and -40 (5*8) because then if any of the first 5 bytes are nonzero, it will automatically be larger and we will instantly know this nonce isn't good enough. For example in the image above, the target would be the second number. The first hash would be too large, and hence fail. The second one is smaller, so it works. 

Obviously, it doesn't have to be the first 5 bytes. That value is up for our choosing, and that is where the concept of *difficulty* comes in- a larger value would indicate a higher difficulty since the hashes have to be even smaller to pass. 

The concept of having to mine each block is called **Proof of Work**, basically we can have confidence in our records because each block was computationally verified. This obviously is very useful for official records like financial balances and transactions, hence why blockchain is so closely tied to cryptocurrency.

Blockchains also have **persistance**, meaning a database of some sort. We use [BoltDB](https://github.com/boltdb/bolt) in this implementation, a key-value store written in Go. Values are stored in **buckets**, and we will use two kinds of key -> value pairs (a simplified version of real Bitcoin implementation). They are:

> 32-byte block-hash -> Block structure (serialized)
> 'l' -> the hash of the last block in a chain

There's also this important concept in crypto of the public/private key pair. Using elliptic curves, we can generate really random numbers, so much so that there are more possiblities than there are atoms in the universe, so the chances of getting the same key pair twice is basically zero. 

![image](https://user-images.githubusercontent.com/69275171/182677005-41d3cb2d-86e7-4eb6-8a51-03bb99fda68a.png)

You might wonder, why does this public/private key stuff matter? Well turns out bitcoin *addresses* are public keys that have been sent through hashing functions (SHA256 and RIPEMD160) and base58 encoded. The address is actually made up of three parts mashed together (see above image)- the version, the actual hash, and the checksum, all base58 encoded. Since hashes are one way, you *cannot* recover the public key from the address. This also means that once you have a pub/priv keypair, you have an address!

# Transactions

Every transaction input in Bitcoin is signed *by the one who created the transaction*. Every transaction in Bitcoin must be verified *before* being put in a block. Verification means (besides other procedures):

1. Checking that inputs have permission to use outputs from previous transactions.
2. Checking that the transaction signature is correct. They will check that: the hash of the public key in an input matches the hash of the referenced output (this ensures that the sender spends only coins belonging to them); the signature is correct (this ensures that the transaction is created by the real owner of the coins).

Bitcoin uses the ECDSA (Elliptic Curve Digital Signature Algorithm) to sign transactions. 

**Digital signature** is an important concept in cryptography, and is just whwen you sign some data with a private key, pass that result to someone, and they decode it with their public key. You might've opened an account before, say with something like Metamask, where you were given a random list of strings. That is your private key. The private key must *never* be leaked because otherwise, people have full access to your account, as they can create digital signatures.

Transactions are hashed, and usually they use what's called a *trimmed copy*, which is a more compact version of the entire Transaction.

# Abbreviations

`TX` = transaction
`UTXO`= "Unspent Transaction Outputs", basically any outputs that have not been referenced by any inputs. We care about these because that is our currency!
`Lock/Unlock`- transaction outputs and inputs are locked by address values. This is how we distinguish who owns which money.
`Genesis Block` = The first block of the blockchain
`Coinbase` = The first transaction on the genesis block
`Wallet` = Nothing more than a public/private key pair
`Address` = The unique string that identifies you on the network. Under the hood, its just a hashed version of your public key.

### Code

Blockchains have blocks (which you can access using Iterator), each block has a list of transactions, and each transaction has a list of inputs/outputs (TXInput, TXOutput). Inputs on the transaction reference previous transactions' outputs, but only one input can correspond to one output and vice versa.

Delete the .db file to create a new blockchain.

Libraries we use
- `encoding/gob` for easy serialization/deserialization
- `boltDB` for persistance
- `flag` for user input
- `github.com/itchyny/base58-go` for base58 encoding (Bitcoin Address generation)
- `big` for large numbers (from ECDSA)