package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/must108/blockchain/blockchain"
)

type CommandLine struct {
	blockchain *blockchain.BlockChain
}

func (cli *CommandLine) printUsage() {
	// prints how you can use this tool
	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println(" print - Prints the blocks in the chain")
}

func (cli *CommandLine) validateArgs() {
	// if the terminal args is less than 2
	if len(os.Args) < 2 {
		cli.printUsage() // prints the instructions
		runtime.Goexit() // exits the application,
		// shuts down the go routine
	}
}

// to add a block... obviously
func (cli *CommandLine) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Added Block!")
}

// to print a blockchain... obviously
func (cli *CommandLine) printChain() {
	iter := cli.blockchain.Iterator() // converts the blockchain
	// to an iterator struct

	for {
		block := iter.Next() // goes to the next block

		// prints the prevHash, curr Block data, and curr Block hash.
		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		// gets the proof of work of the block
		pow := blockchain.NewProof(block)
		// prints the PoW once it is validated.
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		// to break out of the loop
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	// flags, essentially specific strings that if typed into the cli
	// something is run
	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:]) // parses any other arguments after "add"
		blockchain.Handle(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	// .Parsed checks if a flag has been parsed or not
	if addBlockCmd.Parsed() {
		if *addBlockData == "" { // if the pointer is empty
			addBlockCmd.Usage() // print usage
			runtime.Goexit()    // goexit
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func main() {
	defer os.Exit(0)                     // prevents unexpected exit
	chain := blockchain.InitBlockChain() // opens up the database
	defer chain.Database.Close()         // properly close the db before main function end

	cli := CommandLine{chain} // create the cli struct
	cli.run()                 // run the cli struct
}

// a blockchain's hash value is always dependent on the value
// of the previous node.

// if a value is corrupt/wrong, it can easily be identified as
// it will corrupt all further values.
