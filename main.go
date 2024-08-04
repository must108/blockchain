package main

import (
	"fmt"
	"strconv"

	"github.com/must108/blockchain/blockchain"
)

func main() {
	// blockchain initialize

	chain := blockchain.InitBlockChain()

	// adds three blocks after genesis block
	chain.AddBlock("First Block After Genesis")
	chain.AddBlock("Second Block After Genesis")
	chain.AddBlock("Third Block After Genesis")

	// prints prevhash, data, and hash of all blocks
	for _, block := range chain.Blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}

// a blockchain's hash value is always dependent on the value
// of the previous node.

// if a value is corrupt/wrong, it can easily be identified as
// it will corrupt all further values.
