package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

// proof of algorithms / consensus algorithms
// this is a proof of work algorithm

// blockchain is secured by forcing a network to do
// work to add a block to the chain.

// this work is computational power
// blocks and data in blocks are secured through this process

// proof of this work is needed to show that a block is signed

// proof of work steps:
// take data from a block
// create a counter (aka a nonce) that starts at 0
// create a hash of the data + counter
// check the hash against a set of requirements

// set of requirements:
// first few bytes must contain 0

const diff = 18 // static difficulty
// however, in a genuine blockchain, difficulty increments over time

type ProofOfWork struct {
	Block  *Block   // a specific block
	Target *big.Int // a value that determines the
	// validity of a block. based on the diff value
}

func NewProof(b *Block) *ProofOfWork {
	// get a new BigInt
	target := big.NewInt(1)
	// shift number of bytes in target
	// by uint(256-diff) (256 is the number of bytes in our hash)
	// lsh basically just does a left shift
	target.Lsh(target, uint(256-diff))

	// creates and returns a new proof of work
	pow := &ProofOfWork{b, target}
	return pow
}

// replaces the derive hash.
// integrates the difficulty value into the hash.
func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			// uses prevHash and data again, + nonce and diff in bytes,
			// to create a better hash
			// joins the 4 values
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)), // nonce value in bytes
			ToHex(int64(diff)),  // diff value in bytes
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0 // defines a nonce

	for nonce < math.MaxInt64 {
		// essentially an infinite loop
		data := pow.InitData(nonce) // concats our data
		hash = sha256.Sum256(data)  // hashes our concatted data

		fmt.Printf("\r%x", hash)
		intHash.SetBytes((hash[:])) // sets intHash to the full slice of hashes

		if intHash.Cmp(pow.Target) == -1 { // if intHash is less than Target
			break // break as we've found a valid hash for the block
		} else {
			nonce++ // continue with a higher nonce until we find a good val
		}
	}

	fmt.Println()

	return nonce, hash[:] // returns our nonce value and slice of hash
}

// validates a block
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)
	// initializes the data by combining prev hash with nonce, curr data, and diff

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])
	// hashes the data and converts the hash (a slice of bytes) to a bigint

	return intHash.Cmp(pow.Target) == -1
	// returns true if intHash is less than target.
	// otherwise false
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	// takes our number, and converts it into bytes
	err := binary.Write(buff, binary.BigEndian, num)
	// BigEndian specifies the organization of our bytes

	// if err has a value
	Handle(err)

	// return the bytes in buff
	return buff.Bytes()
}
