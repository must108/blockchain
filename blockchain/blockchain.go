package blockchain

import (
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
		lastHash, err = item.Value()

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

func (chain *BlockChain) AddBlock(data string) {
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

	newBlock := CreateBlock(data, lastHash)
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
