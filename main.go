package main

// we want 24 zero bits or more to accept a hash as OK
const targetBits = 16

func main() {

	// startup a cli instance with this blockchain we made last line
	cli := CLI{}

	// parse user input
	cli.Run()
}
