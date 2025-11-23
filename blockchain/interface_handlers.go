package blockchain

import (
	"capychain/dbg"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupInterfaceHandlers(r *mux.Router) {
	r.HandleFunc(
		"/interface",
		interfaceHandler,
	).Methods("GET")
}

func interfaceHandler(w http.ResponseWriter, r *http.Request) {
	dt := CapyBlockchainInstance.Database.NewQueries()
	ctx := r.Context()
	allBlocks, err := dt.GetAllBlocks(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve blocks", http.StatusInternalServerError)
		dbg.Errorf("Error retrieving blocks: %s", err.Error())
		return
	}

	w.Write([]byte("All Blocks:\n"))
	for _, block := range allBlocks {
		w.Write([]byte(
			fmt.Sprintf("Height: %d, Hash: %s, Previous Hash: %s, Timestamp: %d, Nonce: %d, Difficulty: %d, Data: %s\n",
				block.Height,
				block.Hash,
				block.PreviousHash,
				block.Timestamp,
				block.Nonce,
				block.Difficulty,
				block.Data,
			),
		))
	}
}
