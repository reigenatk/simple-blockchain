package main

// we want 24 zero bits or more to accept a hash as OK
const targetBits = 16

func main() {

	// create a blockchain (which writes to db as well)
	blockchain := InitBlockchain()

	// close the DB after we're done
	defer blockchain.DB.Close()

	// startup a cli instance with this blockchain we made last line
	cli := CLI{blockchain}

	// parse user input
	cli.Run()
}
