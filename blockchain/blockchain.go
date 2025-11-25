package blockchain

import (
	"capychain/db"
	"capychain/dbg"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var CapyBlockchainInstance *CapyBlockchain

type CapyBlockchain struct {
	Node     *CapyNode
	Database *CapyDatabase
}

func NewCapyBlockchain(name string, port string, dbSource string) (*CapyBlockchain, error) {
	var err error

	capyNode := NewCapyNode(name, port)

	capyDatabase, err := NewCapyDatabase(dbSource)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %s", err.Error())
	}

	return &CapyBlockchain{
		Node:     capyNode,
		Database: capyDatabase,
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

	if CapyBlockchainInstance.Length() == 0 {
		dbg.Infof("Creating genesis block")
		genesisBlock := NewGenesisCapyBlock("Genesis Block", 0)
		err = CapyBlockchainInstance.AddBlockToDatabase(genesisBlock)
		if err != nil {
			return fmt.Errorf("failed to create genesis block: %s", err.Error())
		}
		dbg.Infof("Genesis block created successfully")
	}

	err = CapyBlockchainInstance.Node.StartServer()
	if err != nil {
		return fmt.Errorf("failed to start API server: %s", err.Error())
	}

	return nil

}

func CalculateCapyBlockHash(block *CapyBlock) string {
	record := fmt.Sprintf(
		"%d%s%d%d%s",
		block.Height,
		block.PreviousHash,
		block.Timestamp,
		block.Nonce,
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

func ValidateCapyBlockBlockchain(capyBlocks []*CapyBlock) (*CapyBlock, error) {
	hashMap := make(map[string]*CapyBlock, len(capyBlocks))
	for _, b := range capyBlocks {
		hashMap[b.Hash] = b
	}

	validated := make(map[string]bool, len(capyBlocks))

	for _, start := range capyBlocks {

		if validated[start.Hash] {
			continue
		}

		cur := start
		visited := make(map[string]bool)

		for {
			if visited[cur.Hash] {
				return cur, fmt.Errorf("cycle detected at hash %s (height %d)", cur.Hash, cur.Height)
			}
			visited[cur.Hash] = true

			if validated[cur.Hash] {
				break
			}

			// Dois critérios aceitos para o genesis block:
			// 1) Height 0
			// 2) PreviousHash == "0"
			if cur.Height == 0 || cur.PreviousHash == "0" {

				// recalcula hash
				if CalculateCapyBlockHash(cur) != cur.Hash {
					return cur, fmt.Errorf("invalid genesis block hash at height %d", cur.Height)
				}

				validated[cur.Hash] = true
				break
			}

			//   BLOCO NORMAL
			prev, ok := hashMap[cur.PreviousHash]
			if !ok {
				return cur, fmt.Errorf("previous block not found for block hash %s (height %d)", cur.Hash, cur.Height)
			}

			if !ValidateBlock(cur, prev) {
				return cur, fmt.Errorf("invalid block at height %d (hash %s)", cur.Height, cur.Hash)
			}

			validated[cur.Hash] = true
			cur = prev
		}
	}

	return nil, nil
}

func (cb *CapyBlockchain) ValidateCapyBlocksBlockchain() (*CapyBlock, error) {
	/*
		Primeiro obtemos os blocos ordenados por height
		e para cada um adicionamos um mark:
			type MarkedBlock struct {
				Mark  bool
				Block *CapyBlock
			}
		Depois iteramos do maior height para o menor,
		validando cada bloco com seu previousHash.
		Ao passarmos por um bloco, o marcamos como validado (Mark = true)
		para evitar validações repetidas.
		Se encontrarmos um bloco inválido, retornamos erro imediatamente.
		Se chegarmos ao genesis block sem erros apartir de todos os blocos, então
		toda a blockchain está válida.
	*/
	var err error
	dt := CapyBlockchainInstance.Database.NewQueries()
	ctx := context.Background()

	var allBlockssOrderedByHeight []db.Block
	allBlockssOrderedByHeight, err = dt.GetAllBlocksOrderedByHeight(ctx)
	if err != nil {
		dbg.Errorf("Error retrieving blocks from database: %s", err.Error())
		return nil, err
	}

	capyBlocks := make([]*CapyBlock, len(allBlockssOrderedByHeight))
	for i, dbBlock := range allBlockssOrderedByHeight {
		capyBlocks[i] = NewCapyBlockFromDbBlock(dbBlock)
	}

	return ValidateCapyBlockBlockchain(capyBlocks)

}

func (cb *CapyBlockchain) MineCapyBlock(data string, resChan chan CapyBlock) {
	difficulty := cb.Node.GetMostVotedNodeMiningDifficulty()

	highestBlock, err := cb.GetHighestBlock()
	if err != nil {
		dbg.Errorf("Error getting highest block: %s", err.Error())
		return
	}

	newBlock := NewCapyBlock(
		highestBlock.Height+1,
		highestBlock.Hash,
		time.Now().Unix(),
		0,
		data,
	)

	start := time.Now().Unix()
	for {
		newBlock.Hash = CalculateCapyBlockHash(newBlock)
		target := strings.Repeat("0", difficulty)
		if strings.HasPrefix(newBlock.Hash, target) {
			elapsed := time.Now().Unix() - start
			dbg.Infof("Block mined: %s in %d seconds", newBlock.Hash, elapsed)
			resChan <- *newBlock
			return
		}
		newBlock.Nonce++
	}
}

func (cb *CapyBlockchain) Length() int64 {
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

func (cb *CapyBlockchain) GetAllCapyBlocks() ([]*CapyBlock, error) {
	var err error
	dt := cb.Database.NewQueries()
	ctx := context.Background()
	dbBlocks, err := dt.GetAllBlocks(ctx)
	if err != nil {
		return nil, err
	}

	capyBlocks := make([]*CapyBlock, len(dbBlocks))
	for i, dbBlock := range dbBlocks {
		capyBlocks[i] = NewCapyBlockFromDbBlock(dbBlock)
	}
	return capyBlocks, nil
}

func (cb *CapyBlockchain) SyncBlockchain() {
	/*
		Strategy implemented:
		- Para cada peer:
			1) Buscar a chain do peer via HTTP (endpoint /blocks).
			2) Deserializar em []CapyBlock.
			3) Validar a chain do peer.
			4) Se válida e maior que a chain local:
				- Encontrar ponto de fork (primeiro height onde os hashes divergem).
				- Deletar blocos locais a partir do ponto de fork.
				- Inserir blocos do peer a partir do ponto de fork.
			- Caso contrário, não alterar a DB local.
		- Continua com os próximos peers mesmo que um falhe.
	*/
	for _, peer := range cb.Node.Peers {
		dbg.Infof("Synchronizing blockchain with peer [%s:%s]", peer.Address, peer.Port)

		resp, err := http.Get(
			fmt.Sprintf("http://%s:%s/chain", peer.Address, peer.Port),
		)
		if err != nil {
			dbg.Errorf("Error fetching chain from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
			continue
		}
		if resp.StatusCode != http.StatusOK {
			dbg.Errorf("Non-OK HTTP status from peer [%s:%s]: %s", peer.Address, peer.Port, resp.Status)
			resp.Body.Close()
			continue
		}

		var peerChain []CapyBlock
		err = json.NewDecoder(resp.Body).Decode(&peerChain)
		if err != nil {
			dbg.Errorf("Error decoding chain from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		tempChain, err := cb.GetAllCapyBlocks()
		if err != nil {
			dbg.Errorf("Error retrieving local chain for synchronization with peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
			continue
		}
		for _, peerCapyBlock := range peerChain {
			tempChain = append(tempChain, &peerCapyBlock)
		}

		invalidBlock, err := ValidateCapyBlockBlockchain(tempChain)
		if err != nil {
			dbg.Errorf("Invalid blockchain from peer [%s:%s] at block height %d, and hash [%s]: %s", peer.Address, peer.Port, invalidBlock.Height, invalidBlock.Hash, err.Error())
			continue
		}

		//
		dt := cb.Database.NewQueries()
		ctx := context.Background()
		for _, tmpBlck := range tempChain {
			err = dt.InsertBlock(
				ctx,
				db.InsertBlockParams{
					Height:       tmpBlck.Height,
					Hash:         tmpBlck.Hash,
					PreviousHash: tmpBlck.PreviousHash,
					Timestamp:    tmpBlck.Timestamp,
					Nonce:        tmpBlck.Nonce,
					Data:         tmpBlck.Data,
				},
			)
			if err != nil {
				// Nothing to do, just continue
			}
		}

		dbg.Infof("Successfully synchronized blockchain with peer [%s:%s]", peer.Address, peer.Port)
	}
}

func (cb *CapyBlockchain) GetCapyBlocksWithMinHeight(minHeight int64) ([]CapyBlock, error) {
	var err error
	dt := cb.Database.NewQueries()
	ctx := context.Background()
	dbBlocks, err := dt.GetBlocksWithMinHeight(ctx, minHeight)
	if err != nil {
		return nil, err
	}

	capyBlocks := make([]CapyBlock, len(dbBlocks))
	for i, dbBlock := range dbBlocks {
		capyBlocks[i] = *NewCapyBlockFromDbBlock(dbBlock)
	}
	return capyBlocks, nil
}

func (cb *CapyBlockchain) GetHighestBlock() (*CapyBlock, error) {
	var err error
	dt := cb.Database.NewQueries()
	ctx := context.Background()
	dbBlock, err := dt.GetHighestBlock(ctx)
	if err != nil {
		return nil, err
	}
	capyBlock := NewCapyBlockFromDbBlock(dbBlock)
	return capyBlock, nil
}

func (cb *CapyBlockchain) Sync() {
	cb.Node.SyncNodePeers()
	cb.SyncBlockchain()
}
