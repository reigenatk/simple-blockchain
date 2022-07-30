## Writing simple blockchain in Go

Following [this](https://jeiwan.net/posts/building-blockchain-in-go-part-1/) I will attempt to write a simple blockchain. 

### Theory

Bitcoin uses **Hashcash** as a hashing protocol. It works by taking the block header (which consists of Timestamp, previous hash and data) and then adding a nonce to it, merging all that into a byte array, and taking the hash. If the hash meets a requirement as put forth by the current difficulty value, it will pass. Otherwise increment the nonce and try again. 

In this implementation we use the SHA256 algorithm, which takes in any input and outputs 256 bytes. A hashing algorithm has a few important properties, the main one is that given the output of the hash, it is nearly impossible to figure out the input. The formal term is **one-way**.

The **nonce** is just like the "guess", and the goal of the miners is to find a working nonce. This might take a lot of guesses depending on the difficulty, which is why miners have to have good hardware specs.

The **difficulty** is another way of saying that we set a certain target. For instance:
> The first three bytes of the hash must be 0

Then when we want to **verify** whether a given hash passes the target, we can check whether that hash is less than `1<<(256 - 3)`, 256 because of SHA256 and -3 because then if any of the first 3 digits are nonzero, it will automatically be larger and we know the hash isn't good enough. Obviously, this 3 value is up for our choosing, and that is where the concept of *difficulty* comes in- a larger value would indicate a higher difficulty since the hashes have to be even smaller to pass.

