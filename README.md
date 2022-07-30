## Writing simple blockchain in Go

Following [this](https://jeiwan.net/posts/building-blockchain-in-go-part-1/) I will attempt to write a simple blockchain. 

### Theory

Bitcoin uses Hashcash as a hashing protocol. It works by taking the block header and then adding a counter to it, then taking the hash. If the hash meets a requirement as put forth by the current difficulty value, it will pass. Otherwise increment the counter and try again.