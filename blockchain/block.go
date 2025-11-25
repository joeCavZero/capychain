package blockchain

import "capychain/db"

type CapyBlock struct {
	Height       int64  `json:"height"`
	Hash         string `json:"hash"`
	PreviousHash string `json:"previous_hash"`
	Timestamp    int64  `json:"timestamp"`
	Nonce        int64  `json:"nonce"`
	Data         string `json:"data"`
}

func NewCapyBlock(height int64, previousHash string, timestamp int64, nonce int64, data string) *CapyBlock {
	return &CapyBlock{
		Height:       height,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
		Nonce:        nonce,
		Data:         data,
	}
}

func NewCapyBlockFromDbBlock(dbBlock db.Block) *CapyBlock {
	return &CapyBlock{
		Height:       dbBlock.Height,
		Hash:         dbBlock.Hash,
		PreviousHash: dbBlock.PreviousHash,
		Timestamp:    dbBlock.Timestamp,
		Nonce:        dbBlock.Nonce,
		Data:         dbBlock.Data,
	}
}

func (cb *CapyBlock) ToDbBlock() db.Block {
	return db.Block{
		Height:       cb.Height,
		Hash:         cb.Hash,
		PreviousHash: cb.PreviousHash,
		Timestamp:    cb.Timestamp,
		Nonce:        cb.Nonce,
		Data:         cb.Data,
	}
}

func (cb *CapyBlock) IsGenesisBlock() bool {
	return cb.Height == 0 && cb.PreviousHash == "0"
}

func NewGenesisCapyBlock(data string, difficulty int64) *CapyBlock {
	genesisBlock := &CapyBlock{
		Height:       0,
		PreviousHash: "0",
		Timestamp:    0,
		Nonce:        0,
		Data:         data,
	}
	genesisBlock.Hash = CalculateCapyBlockHash(genesisBlock)
	return genesisBlock
}
