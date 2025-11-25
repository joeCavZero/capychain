package blockchain

import (
	"capychain/dbg"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupBlockchainHandlers(r *mux.Router) {

	// Endpoint para obter a blockchain completa
	r.HandleFunc(
		"/chain",
		chainGetHandler,
	).Methods("GET")
	// ou a partir de uma altura mínima fornecida
	r.HandleFunc(
		"/chain",
		chainPostHandler,
	).Methods("POST")

	// Endpoint para saber o tamanho da blockchain
	r.HandleFunc(
		"/chain/length",
		chainLengthHandler,
	).Methods("GET")

	// Endpoint para sincronizar a blockchain com outros nós
	r.HandleFunc(
		"/chain/sync",
		chainSyncGetHandler,
	).Methods("GET")

	// Endpoint para iniciar o processo de mineração
	r.HandleFunc(
		"/chain/mine",
		chainMineHandler,
	).Methods("POST")

	// Endpoint para adicionar um novo bloco com dados fornecidos
	r.HandleFunc(
		"/chain/block",
		chainInsertBlockHandler,
	).Methods("POST")

	// Endpoint para deletar um bloco específico
	r.HandleFunc(
		"/chain/block",
		chainDeleteBlockHandler,
	).Methods("DELETE")

	// Endpoint para validar
	r.HandleFunc(
		"/chain/validate",
		chainValidateHandler,
	).Methods("GET")

	// Sync all
	r.HandleFunc(
		"/sync",
		syncHandler,
	).Methods("GET")
}

func chainGetHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var allCapyBlocks []*CapyBlock
	allCapyBlocks, err = CapyBlockchainInstance.GetAllCapyBlocks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		dbg.Errorf("Error retrieving blockchain: %s", err.Error())
		return
	}
	jsonAllCapyBlocks, err := json.Marshal(allCapyBlocks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		dbg.Errorf("Error marshaling blockchain: %s", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonAllCapyBlocks)
}

func chainPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var requestData struct {
		Height int64 `json:"height"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		dbg.Errorf("Error decoding request data: %s", err.Error())
		return
	}

	var allCapyBlocks []CapyBlock
	allCapyBlocks, err = CapyBlockchainInstance.GetCapyBlocksWithMinHeight(requestData.Height)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		dbg.Errorf("Error retrieving blockchain: %s", err.Error())
		return
	}
	jsonAllCapyBlocks, err := json.Marshal(allCapyBlocks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		dbg.Errorf("Error marshaling blockchain: %s", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonAllCapyBlocks)
}

func chainMineHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var requestData struct {
		Data string `json:"data"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		dbg.Errorf("Error decoding request data: %s", err.Error())
		return
	}

	data := requestData.Data

	var resChan chan CapyBlock = make(chan CapyBlock)

	go CapyBlockchainInstance.MineCapyBlock(data, resChan)
	minedBlock := <-resChan

	// persistir o bloco minerado no banco de dados
	CapyBlockchainInstance.AddBlockToDatabase(&minedBlock)

	jsonMinedBlock, err := json.Marshal(minedBlock)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		dbg.Errorf("Error marshaling mined block: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonMinedBlock)
}

func chainValidateHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	dbg.Infof("Starting blockchain validation process")

	blocksLength := CapyBlockchainInstance.Length()

	if blocksLength <= 0 {
		dbg.Warnf("Blockchain is empty (considered valid)")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"is_valid":true}`))
		return
	}

	var inconsistentCapyBlock *CapyBlock
	inconsistentCapyBlock, err = CapyBlockchainInstance.ValidateCapyBlocksBlockchain()

	if err != nil {
		if inconsistentCapyBlock != nil {
			type InconsistentBlockResponse struct {
				IsValid bool      `json:"is_valid"`
				Block   CapyBlock `json:"inconsistent_block"`
			}
			response := InconsistentBlockResponse{
				IsValid: false,
				Block:   *inconsistentCapyBlock,
			}
			jsonResp, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResp)
			dbg.Warnf("Blockchain is invalid at block height %d", inconsistentCapyBlock.Height)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"is_valid":false}`))
			dbg.Warnf("Blockchain is invalid")
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"is_valid":true}`))
		dbg.Infof("Blockchain is valid")
	}
}

func chainInsertBlockHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var capyBlock CapyBlock
	err = json.NewDecoder(r.Body).Decode(&capyBlock)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		dbg.Errorf("Error decoding block data: %s", err.Error())
		return
	}

	err = CapyBlockchainInstance.AddBlockToDatabase(&capyBlock)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		dbg.Errorf("Error inserting block: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func chainSyncGetHandler(w http.ResponseWriter, r *http.Request) {
	CapyBlockchainInstance.SyncBlockchain()
	w.WriteHeader(http.StatusOK)
}

func chainLengthHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	blocksLength := CapyBlockchainInstance.Length()

	jsonResp, err := json.Marshal(
		map[string]int64{
			"length": blocksLength,
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		dbg.Errorf("Error marshaling blockchain length: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func chainDeleteBlockHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	type RequestData struct {
		Height int64  `json:"height"`
		Hash   string `json:"hash"`
	}

	var requestData RequestData
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		dbg.Errorf("Error decoding request data: %s", err.Error())
		return
	}

	err = CapyBlockchainInstance.DeleteBlockByHeightAndHash(requestData.Height, requestData.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		dbg.Errorf("Error deleting block: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	CapyBlockchainInstance.Sync()
	w.WriteHeader(http.StatusOK)
}
