package blockchain

import (
	"capychain/db"
	"capychain/dbg"
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

var CapyBlockchainInstance *CapyBlockchain

type CapyBlockchain struct {
	Node       *CapyNode     `json:"node"`
	Database   *CapyDatabase `json:"database"`
	Difficulty int64         `json:"difficulty"`
}

func NewCapyBlockchain(name string, port string, dbSource string) (*CapyBlockchain, error) {
	var err error

	capyNode := NewCapyNode(name, port)

	capyDatabase, err := NewCapyDatabase(dbSource)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %s", err.Error())
	}

	return &CapyBlockchain{
		Node:       capyNode,
		Database:   capyDatabase,
		Difficulty: 0,
	}, nil
}

func Init(name string, port string, dbSource string) error {
	var err error

	if CapyBlockchainInstance != nil {
		return fmt.Errorf("blockchain instance already initialized")
	}

	CapyBlockchainInstance, err = NewCapyBlockchain(name, port, dbSource)
	if err != nil {
		return fmt.Errorf("failed to initialize blockchain: %s", err.Error())
	}

	if CapyBlockchainInstance.Lenght() == 0 {
		dbg.Infof("Creating genesis block")
		genesisBlock := NewGenesisCapyBlock("Genesis Block", CapyBlockchainInstance.Difficulty)
		err = CapyBlockchainInstance.AddBlockToDatabase(genesisBlock)
		if err != nil {
			return fmt.Errorf("failed to create genesis block: %s", err.Error())
		}
		dbg.Infof("Genesis block created successfully")
	}

	go CapyBlockchainInstance.Node.StartDiscoveryListener()

	err = CapyBlockchainInstance.Node.StartServer()
	if err != nil {
		return fmt.Errorf("failed to start API server: %s", err.Error())
	}

	return nil

}

func CalculateCapyBlockHash(block *CapyBlock) string {
	record := fmt.Sprintf(
		"%d%s%d%d%d%s",
		block.Height,
		block.PreviousHash,
		block.Timestamp,
		block.Nonce,
		block.Difficulty,
		block.Data,
	)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return fmt.Sprintf("%x", hashed)
}

func ValidateBlock(block *CapyBlock, previousBlock *CapyBlock) bool {

	if block.PreviousHash != previousBlock.Hash {
		return false
	}
	if CalculateCapyBlockHash(block) != block.Hash {
		return false
	}
	return true
}

func (cb *CapyBlockchain) ValidateCapyBlocksBlockchain() (*CapyBlock, error) {
	var err error
	dt := CapyBlockchainInstance.Database.NewQueries()
	ctx := context.Background()

	var allBlockssOrderedByHeight []db.Block
	allBlockssOrderedByHeight, err = dt.GetBlocksOrderedByHeight(ctx)
	if err != nil {
		dbg.Errorf("Error retrieving blocks from database: %s", err.Error())
		return nil, err
	}

	type MarkedBlock struct {
		Mark  bool
		Block *CapyBlock
	}

	markedBlocks := make([]MarkedBlock, len(allBlockssOrderedByHeight))
	for i, block := range allBlockssOrderedByHeight {
		markedBlocks[i] = MarkedBlock{
			Mark:  false,
			Block: NewCapyBlockFromDbBlock(block),
		}
	}
	/*
		Iterar em ordem de height maior para menor
		marcando os que já foram validados (Mark = true)
		para evitar validações repetidas
		Sempre checando se o bloco atual tem previousHash até
		chegar no genesis block
		Lembrando que o previousHash sempre tem um height menor
		que o atual bloco da iteração
	*/
	for i := len(markedBlocks) - 1; i >= 0; i-- {
		if markedBlocks[i].Mark {
			continue
		}
		currentBlock := markedBlocks[i].Block
		if currentBlock.IsGenesisBlock() {
			markedBlocks[i].Mark = true
			continue
		}
		previousHash := currentBlock.PreviousHash
		foundPrevious := false
		for j := i - 1; j >= 0; j-- {
			if markedBlocks[j].Block.Hash == previousHash {
				foundPrevious = true
				if ValidateBlock(currentBlock, markedBlocks[j].Block) {
					markedBlocks[i].Mark = true
					markedBlocks[j].Mark = true
				} else {
					dbg.Errorf("Invalid block at height %d", currentBlock.Height)
					return currentBlock, fmt.Errorf("invalid block at height %d", currentBlock.Height)
				}
				break
			}
		}
		if !foundPrevious {
			dbg.Errorf("Previous block not found for block at height %d", currentBlock.Height)
			return currentBlock, fmt.Errorf("previous block not found for block at height %d", currentBlock.Height)
		}
	}

	return nil, nil
}

func (cb *CapyBlockchain) GetLatestBlock() (*CapyBlock, error) {
	dt := cb.Database.NewQueries()
	ctx := context.Background()
	dbBlocks, err := dt.GetAllBlocks(ctx)
	if err != nil {
		return nil, err
	}

	if len(dbBlocks) == 0 {
		return nil, fmt.Errorf("no blocks found in the database")
	}

	latestDbBlock := dbBlocks[len(dbBlocks)-1]
	latestCapyBlock := NewCapyBlockFromDbBlock(latestDbBlock)
	return latestCapyBlock, nil
}

func (cb *CapyBlockchain) MineCapyBlock(data string, resChan chan CapyBlock) {
	lastBlock, err := cb.GetLatestBlock()
	if err != nil {
		dbg.Errorf("Error getting latest block: %s", err.Error())
		return
	}
	newBlock := NewCapyBlock(
		lastBlock.Height+1,
		lastBlock.Hash,
		time.Now().Unix(),
		0,
		lastBlock.Difficulty,
		data,
	)

	start := time.Now().Unix()
	for {
		newBlock.Hash = CalculateCapyBlockHash(newBlock)
		target := strings.Repeat("0", int(newBlock.Difficulty))
		if strings.HasPrefix(newBlock.Hash, target) {
			elapsed := time.Now().Unix() - start
			dbg.Infof("Block mined: %s in %d seconds", newBlock.Hash, elapsed)
			resChan <- *newBlock
			return
		}
		newBlock.Nonce++
	}
}

func (cb *CapyBlockchain) Lenght() int64 {
	dt := cb.Database.NewQueries()
	ctx := context.Background()
	length, err := dt.GetBlocksLength(ctx)
	if err != nil {
		dbg.Errorf("Error getting blocks length: %s", err.Error())
		return 0
	}
	return length
}

func (cb *CapyBlockchain) AddBlockToDatabase(block *CapyBlock) error {
	var err error

	dt := cb.Database.NewQueries()
	ctx := context.Background()
	err = dt.InsertBlock(
		ctx,
		db.InsertBlockParams{
			Height:       block.Height,
			Hash:         block.Hash,
			PreviousHash: block.PreviousHash,
			Timestamp:    block.Timestamp,
			Nonce:        block.Nonce,
			Difficulty:   block.Difficulty,
			Data:         block.Data,
		},
	)
	return err
}

func (cb *CapyBlockchain) DeleteBlockByHeightAndHash(height int64, hash string) error {
	var err error

	dt := cb.Database.NewQueries()
	ctx := context.Background()
	err = dt.DeleteBlockByHeightAndHash(
		ctx,
		db.DeleteBlockByHeightAndHashParams{
			Height: height,
			Hash:   hash,
		},
	)
	return err
}

func (cb *CapyBlockchain) GetAllCapyBlocks() ([]CapyBlock, error) {
	var err error
	dt := cb.Database.NewQueries()
	ctx := context.Background()
	dbBlocks, err := dt.GetAllBlocks(ctx)
	if err != nil {
		return nil, err
	}

	capyBlocks := make([]CapyBlock, len(dbBlocks))
	for i, dbBlock := range dbBlocks {
		capyBlocks[i] = *NewCapyBlockFromDbBlock(dbBlock)
	}
	return capyBlocks, nil
}
