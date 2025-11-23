package blockchain

import (
	"capychain/db"
	"capychain/dbg"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

	if CapyBlockchainInstance.Length() == 0 {
		dbg.Infof("Creating genesis block")
		genesisBlock := NewGenesisCapyBlock("Genesis Block", CapyBlockchainInstance.Difficulty)
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
	allBlockssOrderedByHeight, err = dt.GetAllBlocksOrderedByHeight(ctx)
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

func (cb *CapyBlockchain) MineCapyBlock(data string, resChan chan CapyBlock) {
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
		highestBlock.Difficulty,
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

func (cb *CapyBlockchain) SynchronizeBlockchain() error {
	for _, peer := range cb.Node.Peers {
		dbg.Infof("Synchronizing blockchain with peer [%s:%s]", peer.Address, peer.Port)

		// --- GET chain length do peer
		resp, err := http.Get(fmt.Sprintf("http://%s:%s/chain/length", peer.Address, peer.Port))
		if err != nil {
			// resp pode ser nil aqui, não feche
			return fmt.Errorf("error fetching chain length from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
		}
		if resp != nil {
			defer resp.Body.Close()
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("non-OK HTTP status from peer [%s:%s]: %s", peer.Address, peer.Port, resp.Status)
		}

		var peerChainLengthResp struct {
			Length int64 `json:"length"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&peerChainLengthResp); err != nil {
			return fmt.Errorf("error decoding chain length from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
		}
		// resp.Body já será fechado pelo defer

		localChainLength := cb.Length()
		if peerChainLengthResp.Length <= localChainLength {
			dbg.Infof("Local blockchain is up-to-date with peer [%s:%s]", peer.Address, peer.Port)
			continue
		}

		// --- Solicitar blocos a partir de localChainLength + 1
		var chainPostReqBody struct {
			Height int64 `json:"height"`
		}
		chainPostReqBody.Height = localChainLength + 1

		jsonReqBody, err := json.Marshal(chainPostReqBody)
		if err != nil {
			return fmt.Errorf("error marshaling chain request body for peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
		}

		resp2, err := http.Post(
			fmt.Sprintf("http://%s:%s/chain", peer.Address, peer.Port),
			"application/json",
			strings.NewReader(string(jsonReqBody)),
		)
		if err != nil {
			if resp2 != nil {
				resp2.Body.Close()
			}
			return fmt.Errorf("error fetching chain from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
		}
		if resp2 != nil {
			defer resp2.Body.Close()
		}

		if resp2.StatusCode != http.StatusOK {
			return fmt.Errorf("non-OK HTTP status from peer [%s:%s]: %s", peer.Address, peer.Port, resp2.Status)
		}

		var peerBlocks []CapyBlock
		if err := json.NewDecoder(resp2.Body).Decode(&peerBlocks); err != nil {
			return fmt.Errorf("error decoding chain from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
		}

		// --- Ordenar peerBlocks por Height asc (precaução)
		sort.Slice(peerBlocks, func(i, j int) bool {
			return peerBlocks[i].Height < peerBlocks[j].Height
		})

		// --- Validar que o primeiro bloco recebido conecta ao nosso highest block
		localHighest, err := cb.GetHighestBlock()
		if err != nil {
			return fmt.Errorf("error getting local highest block before sync with peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
		}

		// Se não houver blocos locais (apenas genesis ausente), trate conforme sua regra. Aqui assumimos que localHighest é sempre válido.
		if len(peerBlocks) == 0 {
			dbg.Infof("Peer [%s:%s] não retornou blocos a partir de height %d", peer.Address, peer.Port, chainPostReqBody.Height)
			continue
		}

		firstPeerBlock := &peerBlocks[0]
		// primeiro bloco deve apontar para o hash do bloco local highest (ou ser genesis válido)
		if firstPeerBlock.PreviousHash != localHighest.Hash {
			return fmt.Errorf("peer [%s:%s] chain does not connect to local chain: expected previous hash %s but got %s at peer block height %d",
				peer.Address, peer.Port, localHighest.Hash, firstPeerBlock.PreviousHash, firstPeerBlock.Height)
		}

		// --- Validar a cadeia inteira (hashes e linking) antes de inserir
		for i := 0; i < len(peerBlocks); i++ {
			current := &peerBlocks[i]

			// recompute hash check
			if CalculateCapyBlockHash(current) != current.Hash {
				return fmt.Errorf("invalid hash for peer block at height %d from peer [%s:%s]", current.Height, peer.Address, peer.Port)
			}

			// check previous hash linking
			if i == 0 {
				// já checado contra localHighest
			} else {
				prev := &peerBlocks[i-1]
				if current.PreviousHash != prev.Hash {
					return fmt.Errorf("peer chain has gap/bad link between heights %d and %d from peer [%s:%s]", prev.Height, current.Height, peer.Address, peer.Port)
				}
				if !ValidateBlock(current, prev) {
					return fmt.Errorf("peer chain failed validation at height %d from peer [%s:%s]", current.Height, peer.Address, peer.Port)
				}
			}
		}

		// --- Inserir sequencialmente (apenas depois de validada)
		for _, b := range peerBlocks {
			if err := cb.AddBlockToDatabase(&b); err != nil {
				// Se falhar por duplicata, você pode optar por ignorar; aqui retornamos o erro
				return fmt.Errorf("error adding block from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
			}
		}

		dbg.Infof("Successfully synchronized blockchain with peer [%s:%s]", peer.Address, peer.Port)
		// resp2.Body será fechado pelo defer
	}

	return nil
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

func (cb *CapyBlockchain) Synchronize() error {
	var err error

	err = cb.SynchronizeBlockchain()
	if err != nil {
		return err
	}

	err = cb.Node.SynchronizePeers()
	if err != nil {
		return err
	}

	return nil
}
