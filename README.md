## Writing simple blockchain in Go

Following [this](https://jeiwan.net/posts/building-blockchain-in-go-part-1/) I will attempt to write a simple blockchain. I know the basics of crypto (for example all the buzzwords) but decided that to *really* understand it, I should probably write up an implementation. After all, that's the best way to really understand something. Also, this implementation will try to follow Bitcoin as much as possible.

> Note this readme is really long, mostly so I can go back and read it when I inevitably forget what I did. 

### Install/Run
Just run `go install` and command is called `blockchain`. Make sure go/bin is inside your PATH.
```
Usage:
  getbalance -address ADDRESS - Get balance of ADDRESS
  newblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS
  printchain - Print all the blocks of the blockchain
  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO
  listaddresses - list all the addresses on this network
  createwallet - Generates a public/private keypair, returns your address
  clear - Clears all the files (blockchain.db) and (wallets.dat)
```

### Concepts

First is the concept of a **Block**, which is merely just the following
```
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}
```

Most important is the `PrevBlockHash`, this is the hash of the previous block and is the "chain" part of blockchain. Without the previous hash, its just a collection of blocks.

Next is `Transactions`, which is a list of transactions on the block. We will get to this in another section. And finally a Hash/Nonce, which has to do with mining. 

# Hashing/Mining

Bitcoin uses **Hashcash** as a hashing protocol. It works by taking the block header (which consists of Timestamp, previous hash and transactions) and then adding a nonce to it, merging all that into a byte array, and taking the hash. If the hash meets a requirement as put forth by the current difficulty value, it will pass. Otherwise increment the nonce and try again. 

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

# Signing Transactions

![image](https://user-images.githubusercontent.com/69275171/182747076-e7c9e386-5bd0-4867-9b26-7e6948a6957d.png)

Another important concept is the **Transaction**. In Bitcoin, transactions have inputs and outputs. The input is kind of like "spending" money, and the "output" is like available money. Each input can map to only one output, but a transaction can have multiple inputs.

Inputs store the public key as well as a signature (which was signed by the private key). Outputs store the hash of the public key of the person the money belongs to.

Every transaction input in Bitcoin is signed *by the one who created the transaction*. Every transaction in Bitcoin must be verified *before* being put in a block. Verification means (besides other procedures):

1. Checking that inputs have permission to use outputs from previous transactions.
2. Checking that the transaction signature is correct. They will check that: the hash of the public key in an input matches the hash of the referenced output (this ensures that the sender spends only coins belonging to them); the signature is correct (this ensures that the transaction is created by the real owner of the coins).

Bitcoin uses the ECDSA (Elliptic Curve Digital Signature Algorithm) to sign transactions. 

**Digital signature** is an important concept in cryptography, and is just when  you sign some data with a private key, pass that result to someone, and they decode it with their public key. You might've opened an account before, say with something like Metamask, where you were given a random list of strings. That is your private key. The private key must *never* be leaked because otherwise, people have full access to your account, as they can create digital signatures.

It wouldn't be too far off to say that Bitcoin's reliability comes from digital signatures. Without it, there would be no security.

Also in Bitcoin, everything is identified by hashes. The transactions themselves are individually hashed (the ID of the transaction is just a hash), put together into a transaction array (which lives in a block). The Transaction array is hashed to a byte slice via the Merkel Tree, and finally the block is serialized (using gob) to get the block hash, which is the "name" of the block you see on blockchain sites like [Blockchain.com](https://www.blockchain.com/explorer?view=btc)

# UTXO Set

Just as we stored blocks in BoltDB, so will we store transactions. The reasoning is simple, right now  we need to find the full list of UTXOs every time we want to do a transaction, which means iterating over **every single block in the blockchain**, since transactions reside in blocks! This is not sustainable at all, bitcoin for instance has 747,836 blocks with about 500 GB of data (as of Aug 2022). This would take too much time to go through.

The solution is we create the **UTXO set**, which acts as a kind of cache of sorts. The idea behind this is the following- not all blocks matter when we care about finding the balance of an address. Only unspent transaction outputs matter since those are the only ones that provide a balance. Therefore the philosophy is to have a set of unspent transaction outputs ready at hand, that way we don't need to scan each block.

We will make another bucket in Bolt called `UTXOSet`, and this bucket will hold the follow key-value pairs

> 32-byte transaction hash -> list of unspent transaction output records for that transaction (stores as a TXOutputs object)

And every time we make a block, we will "update" the UTXOSet, as well as provide a general `reindex` function to rescan the whole blockchain. But we will only call that function upon blockchain initialization, as it is costly. Finally, whenever we send money (thus creating a new block), in order to scan the balance of a person, instead of scanning the whole blockchain we simply use the UTXOSet, saving tons of time, especially if the blockchain is large.

As a cool exercise, check out [this site](https://statoshi.info/d/000000009/unspent-transaction-output-set?orgId=1&refresh=10m). You can see how many tarnsactions there are with unspent outputs, how many UTXOs there are total, the size of the UTXO set, and how many bitcoins exist. About 837 million unspent transactions, make up the entirety of the 19 million bitcoins! Also fun fact, Bitcoin has a hard cap at **21 million**. We're getting close to mining all of it!

# Merkle Tree 

The UTXO set was a good optimization, but when creating it, we still need to go into each block and go through a list of all the transactions. What if there was a way in which we could tell whether a transaction was inside of a block, without looping through the array of transactions in the block? This is where a **Merkle Tree** (also known as hash tree) can help. The name sounds fancy, but it's literally just a tree where parent nodes are made up of the hashes of their child nodes. They allow for logarithmic querying as to whether or not something belongs in the tree, which is better than the other, obvious constant time approach of looping through everything.

![image](https://user-images.githubusercontent.com/69275171/183257263-7e54ff16-9c34-4f36-aa62-4bc759e620e6.png)

It works by taking each transaction, arranging it in a tree-like structure, and repeatedly concatenating and hashing the results until there is just one hash left. Then that hash is put in the block header.

# Network
Bitcoin wouldn't be worth anything without users! And users means there must be a network. Blockchains are peer-to-peer, meaning **there is no central authority!** Each user on the Bitcoin Network is formally called a **node**. Right now, there seems to be about [15,000 nodes connected](https://bitnodes.io/). To become a node, all you have to do is download Bitcoin Core, and run it on your PC!

The three types of nodes are Miners, Full nodes, and SPV nodes. Miners simply try to hash blocks, Full nodes are responsible for node discovery and verifying mined blocks, as well as verifying transaction signatures. And SPVs are kinda like Full nodes except they don't keep a full copy of the blockchain. They also help to verify transactons.

How do users communicate? Can they just send stuff willy nilly to each other? Of course not. There's a standard, of course! There are roughly 20 or so kinds of **message formats** you can send, the full list is listed in section 3 [on this page](https://en.bitcoin.it/wiki/Protocol_documentation). But the ones we will focus on are "version", "getblocks", "addr", "block", "inv", "getdata", and "tx".

The rough idea is that we want to download the full blockchain, if we do not yet have it. This can be achieved using the above message types. Then once we have the full blockchain (think of it as being up to date to your favorite Netflix show!), we can now start talking with other peers about the latest blocks coming into the network in real time. But until then, we cannot participate, since we need to catch up.

Some more important terms, the **mempool** is where transactions go to wait for nodes to verify them. Miners put the transactions into blocks, which then get verified, and that reduces the size of the mempool. Another term is the **height** of a block, this is just which block it is in the entire blockchain.

# Abbreviations

`TX` = transaction
`UTXO`= "Unspent Transaction Outputs", basically any outputs that have not been referenced by any inputs. We care about these because that is our currency!
`Lock/Unlock`- transaction outputs and inputs are locked by address values. This is how we distinguish who owns which money.
`Genesis Block` = The first block of the blockchain
`Coinbase` = The first transaction on the genesis block
`Wallet` = Nothing more than a public/private key pair
`Address` = The unique string that identifies you on the network. Under the hood, its just a hashed version of your public key.
`ECDSA` = Elliptic Curve Digital Signature Algorithm

### Code

Blockchains have blocks (which you can access using Iterator), each block has a list of transactions, and each transaction has a list of inputs/outputs (TXInput, TXOutput). Inputs on the transaction reference previous transactions' outputs, but only one input can correspond to one output and vice versa.

Code structure:
- `cli.go` are the functions that are immediately called after your input in the command line
- `blockchain.go` and `transaction.go` are probably the two most important files for they contain the core logic of how crypto works
- `proofofwork.go` for the mining stuff


Libraries used:
- `encoding/gob` for easy serialization/deserialization
- `boltDB` for persistance
- `flag` for user input
- `github.com/btcsuite/btcutil/base58` for base58 encoding (used for Bitcoin Address generation)
- `elliptic` for elliptic curves
- `ecdsa` for elliptic curve algorithms `sign` and `verify`
- `big` for large numbers (which are created from ECDSA)
- `golang.org/x/crypto/ripemd160` for RIPEMD160 hash algorithm (used for publickey -> address)

### Video

Sorry for bad mic :(
[](https://user-images.githubusercontent.com/69275171/183265546-00440ba1-a6dd-449a-9b28-2992cd9a6d83.mp4)


