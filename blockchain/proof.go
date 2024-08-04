package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
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
	Block  *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	// get a new BigInt
	target := big.NewInt(1)
	// shift number of bytes in target
	// by uint(256-diff) (256 is the number of bytes in our hash)
	// lsh basically just does a left shift
	target.Lsh(target, uint(256-diff))

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
			pow.Block.Data,
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

	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
		intHash.SetBytes((hash[:]))

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}

	fmt.Println()

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	// takes our number, and converts it into bytes
	err := binary.Write(buff, binary.BigEndian, num)
	// BigEndian specifies the organization of our bytes

	// if err has a value
	if err != nil {
		log.Panic(err) // an unexpected error
	}

	// return the bytes in buff
	return buff.Bytes()
}
