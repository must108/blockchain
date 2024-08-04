package blockchain

type BlockChain struct {
	Blocks []*Block // slice of block pointers
}

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte // slice of bytes
	Nonce    int
}

func CreateBlock(data string, prevHash []byte) *Block {
	// creates a new block based on a previous hash and
	// the new block's supposed data.
	block := &Block{[]byte{}, []byte(data), prevHash, 0}

	// gets the proof of work per block
	pow := NewProof(block)

	// returns nonce and hash when pow alg is run
	nonce, hash := pow.Run()

	// saves these values
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (chain *BlockChain) AddBlock(data string) {
	// gets the prevblock
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	// creates a new block based on prevblock hash and ccrr block's
	// supposed data.
	new := CreateBlock(data, prevBlock.Hash)
	// appends the new block to the existing chain
	// of blocks (a blockchain.... :OOOOOO)
	chain.Blocks = append(chain.Blocks, new)
}

func Genesis() *Block {
	// creates an initial "Genesis" block
	// to start the blockchain
	return CreateBlock("Genesis", []byte{})
}

func InitBlockChain() *BlockChain {
	// calls the genesis function to
	// initialize the blockchain
	return &BlockChain{[]*Block{Genesis()}}
}
