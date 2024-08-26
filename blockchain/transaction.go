package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput  // slice of inputs
	Outputs []TxOutput // slice of outputs
}

type TxOutput struct {
	Value  int    // the value
	PubKey string // to access a specific value
}

type TxInput struct {
	ID  []byte // references a transaction
	Out int    // index of the output
	Sig string // provides data, used in output's pubkey
}

// create a hash based on bytes
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	// encode the transaction
	encode := gob.NewEncoder((&encoded))
	err := encode.Encode(tx)

	// handle any potential error
	Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// takes in an address to, and string data
// outputs a pointer to a transaction
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	// txin takes a new TxInput with empty slice of bytes,
	// output index of -1, and the data
	txin := TxInput{[]byte{}, -1, data}

	// txout takes in the reward (100 tokens)
	// and pubkey string, which references to address!
	txout := TxOutput{100, to}

	// nil for id, and pass in TxInput and TxOutput slices
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetID()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	// check if length of inputs is 1
	// check if the first input's id is 0
	// check if the input's out index is -1
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// if the unlock functions return true, it means
// that the account owns the output/ref to output from input

// check if the signature value is the same as the passed in data
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// check if the Pubkey is the same as the passed in data
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}

// genesis block has our first transaction
// known as a coinbase transaction
// reward associated with a coinbase transaction
// awarded to user that mines a specific coinbase
// 8:10
