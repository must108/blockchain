package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte // slice of bytes
	Nonce        int
}

// use hashing to provide unique representation of combined
// transactions
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	// creates a new block based on a previous hash and
	// the new block's supposed data.
	block := &Block{[]byte{}, txs, prevHash, 0}

	// gets the proof of work per block
	pow := NewProof(block)

	// returns nonce and hash when pow alg is run
	nonce, hash := pow.Run()

	// saves these values
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func Genesis(coinbase *Transaction) *Block {
	// creates an initial "Genesis" block
	// to start the blockchain
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

// convert data to slice of bytes
func (b *Block) Serialize() []byte {
	// dynamically growing buffer of bytes
	// useful for building/manipulating strings
	var res bytes.Buffer
	// encoder called on our bytes buffer
	encoder := gob.NewEncoder(&res)

	// encodes our actual block, can return an error
	err := encoder.Encode(b)

	// if error occurs
	Handle(err)

	return res.Bytes() // return the result in bytes
}

// converts data in bytes to a Block
func Deserialize(data []byte) *Block {
	var block Block // store the Block val

	// new decoder called on the bytes, reads from the byte slice
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// decodes data into the block
	err := decoder.Decode(&block)

	Handle(err) // error handling

	return &block // returns decoded block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err) // panic is like console.error()
	}
}
