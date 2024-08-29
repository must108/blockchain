package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/must108/blockchain/blockchain"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	// prints how you can use this tool
	fmt.Println("Usage:")
	fmt.Println("getbalance -address ADDRESS - get the balance for an address")
	fmt.Println("createblockchain -address ADDRESS - creates a blockchain and sends genesis reward to address")
	fmt.Println("printchain - Prints the blocks in the chain")
	fmt.Println("send -from FROM -to TO -amount AMOUNT - Send amount of coins")

}

func (cli *CommandLine) validateArgs() {
	// if the terminal args is less than 2
	if len(os.Args) < 2 {
		cli.printUsage() // prints the instructions
		runtime.Goexit() // exits the application,
		// shuts down the go routine
	}
}

// to print a blockchain... obviously
func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next() // goes to the next block

		// prints the prevHash, curr Block data, and curr Block hash.
		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
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

func (cli *CommandLine) createBlockChain(address string) {
	chain := blockchain.InitBlockChain(address) // init block chain
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string) {
	chain := blockchain.ContinueBlockChain(address) // open the blockchain
	defer chain.Database.Close()                    // defer close of db

	balance := 0                     // create balance
	UTXOs := chain.FindUTXO(address) // get unspent txn outputs

	for _, out := range UTXOs {
		balance += out.Value // iterate thru address, and get output values
	}

	fmt.Printf("Balance of %s: %d\n", address, balance) // print the balance, with address
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain := blockchain.ContinueBlockChain(from)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	// flags, essentially specific strings that if typed into the cli
	// something is run
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	// check flags
	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:]) // parses any other arguments after "add"
		if err != nil {
			log.Panic(err)
		}

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	// .Parsed checks if a flag has been parsed or not
	if getBalanceCmd.Parsed() {
		// checks for valid values
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		// checks for valid values
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		// checks for valid values
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

func main() {
	defer os.Exit(0)     // prevents unexpected exit
	cli := CommandLine{} // create the cli struct
	cli.run()            // run the cli struct
}

// a blockchain's hash value is always dependent on the value
// of the previous node.

// if a value is corrupt/wrong, it can easily be identified as
// it will corrupt all further values.
