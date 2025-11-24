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
		"/mine",
		mineHandler,
	).Methods("POST")

	// Endpoint para adicionar um novo bloco com dados fornecidos
	r.HandleFunc(
		"/block",
		insertBlockHandler,
	).Methods("POST")

	// Endpoint para validar
	r.HandleFunc(
		"/validate",
		validateHandler,
	).Methods("GET")

	r.HandleFunc(
		"/block",
		deleteBlockHandler,
	).Methods("DELETE")
}

func chainGetHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var allCapyBlocks []*CapyBlock
	allCapyBlocks, err = CapyBlockchainInstance.GetAllCapyBlocks()
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to retrieve blockchain",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error retrieving blockchain: %s", err.Error())
		return
	}
	jsonAllCapyBlocks, err := json.Marshal(allCapyBlocks)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to marshal blockchain",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
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
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Invalid request data",
			},
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error decoding request data: %s", err.Error())
		return
	}

	var allCapyBlocks []CapyBlock
	allCapyBlocks, err = CapyBlockchainInstance.GetCapyBlocksWithMinHeight(requestData.Height)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to retrieve blockchain",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error retrieving blockchain: %s", err.Error())
		return
	}
	jsonAllCapyBlocks, err := json.Marshal(allCapyBlocks)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to marshal blockchain",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error marshaling blockchain: %s", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonAllCapyBlocks)
}

func mineHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var requestData struct {
		Data string `json:"data"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Invalid request data",
			},
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
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
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to marshal mined block",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error marshaling mined block: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonMinedBlock)

}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	dbg.Infof("Starting blockchain validation process")

	blocksLength := CapyBlockchainInstance.Length()

	if blocksLength <= 0 {
		jsonResp, _ := json.Marshal(
			map[string]string{
				"message": "Blockchain is empty",
			},
		)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResp)
		return
	}

	var inconsistentCapyBlock *CapyBlock
	inconsistentCapyBlock, err = CapyBlockchainInstance.ValidateCapyBlocksBlockchain()

	if err != nil {
		if inconsistentCapyBlock != nil {
			jsonResp, _ := json.Marshal(
				map[string]any{
					"is_valid":           false,
					"error":              err.Error(),
					"inconsistent_block": inconsistentCapyBlock,
				},
			)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResp)
			dbg.Infof("Blockchain is invalid at block height %d", inconsistentCapyBlock.Height)
		} else {
			jsonResp, _ := json.Marshal(
				map[string]any{
					"is_valid": false,
					"error":    err.Error(),
				},
			)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResp)
			dbg.Infof("Blockchain is invalid")
		}
	} else {
		jsonResp, _ := json.Marshal(
			map[string]any{
				"is_valid": true,
				"message":  "Blockchain is valid",
			},
		)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResp)
		dbg.Infof("Blockchain is valid")
	}
}

func insertBlockHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var capyBlock CapyBlock
	err = json.NewDecoder(r.Body).Decode(&capyBlock)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Invalid block data",
			},
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error decoding block data: %s", err.Error())
		return
	}

	err = CapyBlockchainInstance.AddBlockToDatabase(&capyBlock)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to insert block",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error inserting block: %s", err.Error())
		return
	}

	jsonResp, _ := json.Marshal(
		map[string]string{
			"message": "Block added successfully",
		},
	)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func deleteBlockHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	type RequestData struct {
		Height int64  `json:"height"`
		Hash   string `json:"hash"`
	}

	var requestData RequestData
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Invalid request data",
			},
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error decoding request data: %s", err.Error())
		return
	}

	err = CapyBlockchainInstance.DeleteBlockByHeightAndHash(requestData.Height, requestData.Hash)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to delete block",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error deleting block: %s", err.Error())
		return
	}

	jsonResp, _ := json.Marshal(
		map[string]string{
			"message": "Block deleted successfully",
		},
	)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
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
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to marshal blockchain length",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		dbg.Errorf("Error marshaling blockchain length: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
