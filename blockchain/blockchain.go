package blockchain

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST" // verify blockchain db existence
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB // native golang db
}

type BlockChainIterator struct {
	CurrentHash []byte     // the current hash (obv)
	Database    *badger.DB // pointer to badger db
}

func DBexists() bool {
	// if the db doesnt exist, return false, else true
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func ContinueBlockChain(address string) *BlockChain {
	if DBexists() == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})

		return err
	})
	Handle(err)

	chain := BlockChain{lastHash, db}

	return &chain
}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	// checks if db exists
	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit() // exit if so
	}

	// create badger db
	opts := badger.DefaultOptions(dbPath) // badgerdb default options
	opts.Dir = dbPath                     // where keys and metadata are stored in the db
	opts.ValueDir = dbPath                // where values are stored

	opts.Logger = nil // removes logging, as it cluttered the output

	db, err := badger.Open(opts)
	Handle(err) // handles db errs

	// this func is called a "Closure"
	// txn is a pointer to a badger transaction
	// passes back an err
	err = db.Update(func(txn *badger.Txn) error {
		// checks the transaction for a lasthash key (lh),
		// if not found, will return badget.ErrKeyNotFound
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			// address of this transaction is rewarded
			cbtx := CoinbaseTx(address, genesisData)
			genesis := Genesis(cbtx)
			fmt.Println("Genesis created") // when the genesis is initialized
			// txn Set puts a val into our database
			// in this case, genesis Hash is used as the key,
			// and the serialized genesis value is used as the val.
			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			// the lh key is used to store the genesisHash value
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh")) // gets the  lastHash from the db
			Handle(err)
			err = item.Value(func(val []byte) error {
				lastHash = append([]byte(nil), val...)
				return nil
			})
			// gets the value of the item with the key of lh
			return err
		}
	})

	Handle(err)

	// blockchain created with the lastHash and pointer to db
	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	// read-only type transaction, takes in a closure
	// with the txn pointer. returns an error.
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh")) // get the lastHash from the db
		Handle(err)                        // handle a potential err
		err = item.Value(func(val []byte) error {
			lastHash = append([]byte(nil), val...)
			return nil
		}) // unwrap the value and store it in lasthash

		return err // return a potential error
	})

	Handle(err)

	newBlock := CreateBlock(transactions, lastHash)
	// creates a new block with our data and the lastHash value

	err = chain.Database.Update(func(txn *badger.Txn) error {
		// hash is used as key, serialized newBlock used as value
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash) // set hash val to "lh" key

		chain.LastHash = newBlock.Hash
		// make the lastHash the current hash, for the next block

		return err
	})
	Handle(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	// creates a BlockChainIterator by getting the chain's
	// lastHash and the pointer to its Database
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter // returns this new iterator
} // iterates through the newest block to the genesis

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	// read-only on the db
	err := iter.Database.View(func(txn *badger.Txn) error {
		// gets the item with the currenthash
		item, _ := txn.Get(iter.CurrentHash)
		// get a byte representation of our block + an error
		err := item.Value(func(encodedBlock []byte) error {
			block = Deserialize(encodedBlock)
			return nil
		}) // Deserialize the encodedblock to create a new block

		return err
	})
	Handle(err)

	// keep going backwards by getting the prev hash.
	// like traversing through a linked list
	iter.CurrentHash = block.PrevHash

	return block
}

// unspent transactions - transactions that have outputs not referenced by other inputs

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction // arr of transactions

	spentTXOs := make(map[string][]int) // mapped string key with val of slice int

	iter := chain.Iterator() // iterate thru blockchain

	for {
		block := iter.Next() // get block from db

		for _, tx := range block.Transactions { // iterate through transactions
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs { // iterate thru transaction
				if spentTXOs[txID] != nil { // if the map with a specific txID exists
					for _, spentOut := range spentTXOs[txID] { // iterate thru that map
						if spentOut == outIdx { // is the spentOut index equal to the output index
							continue Outputs // continue with the outputs for loop
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				} // append all transactions that can be unlocked into unspentTxs
			}

			if tx.IsCoinbase() == false { // check if this is a coinbase transaction
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) { // if in can unlock a specific address
						inTxID := hex.EncodeToString(in.ID)           // encode in.ID
						spentTXOs[inTxID] = append(spentTXOs[inTxID]) // append a spent transaction
						// with the encoded id to spentTxos
					}
				}
			}
		}

		if len(block.PrevHash) == 0 { // if len of block prevhash is 0, block is genesis
			break
		}
	}

	return unspentTxs
}

func (chain *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput // array of transaction outputs
	unspentTransactions := chain.FindUnspentTransactions(address)
	// use our method to find unspent transaction

	for _, tx := range unspentTransactions { // iterate thru unspentTransactions
		for _, out := range tx.Outputs { // iterate thru transaction outputs
			if out.CanBeUnlocked(address) { // check if output can be unlocked
				UTXOs = append(UTXOs, out) // if so, append it to UTXOs
			}
		}
	}

	return UTXOs // return UTXOs
}

// enables us to create normal transactions, that are not coinbase transactions
// takes in address and a specified amount, outputs an int and a map w string and int
func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address) // finds unspent txns
	accumulated := 0

Work:
	for _, tx := range unspentTxs { // iterates thru unspent transactions
		txID := hex.EncodeToString(tx.ID) // get tx id. and encode into string

		for outIdx, out := range tx.Outputs { // iterate thru outputs
			if out.CanBeUnlocked(address) && accumulated < amount { // can output be unlocked,
				// and is the accumulated less than the amnt we want to send
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx) // append new outIdx to unspentOuts

				if accumulated >= amount { // if accumulated is greater or equal to amount, break
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts // generate general transaction with these vals
}
