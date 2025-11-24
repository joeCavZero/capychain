package blockchain

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupNodeHandlers(r *mux.Router) {

	// Endpoint para obter informações do nó
	r.HandleFunc(
		"/node",
		nodeHandler,
	).Methods("GET")

	// Endpoint para adicionar um novo peer
	r.HandleFunc(
		"/peers",
		peersPostHandler,
	).Methods("POST")

	// Endpoint para iniciar sincronização de peers
	r.HandleFunc(
		"/node/sync",
		nodeSyncGetHandler,
	).Methods("GET")

	// Endpoint para sincronizar peers
	r.HandleFunc(
		"/node/sync",
		nodeSyncPostHandler,
	).Methods("POST")
}

func nodeHandler(w http.ResponseWriter, r *http.Request) {
	/*
		Write CapyNode information as JSON to response
	*/
	jsonedNode, err := json.Marshal(CapyBlockchainInstance.Node)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to marshal node information",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonedNode)
}

func peersPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var newPeer CapyPeer
	err = json.NewDecoder(r.Body).Decode(&newPeer)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to decode peer information",
			},
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		return
	}

	CapyBlockchainInstance.Node.AddCapyPeer(newPeer)

	jsonedResp, _ := json.Marshal(
		map[string]string{
			"message": "Peer added successfully",
		},
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonedResp)
}

func nodeSyncGetHandler(w http.ResponseWriter, r *http.Request) {
	CapyBlockchainInstance.Node.SyncNodePeers()
	w.WriteHeader(http.StatusOK)
}

func nodeSyncPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var passedCapyNodes []CapyNode
	err = json.NewDecoder(r.Body).Decode(&passedCapyNodes)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to decode array of addresses",
			},
		)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		return
	}

	CapyBlockchainInstance.Node.CastNodePeersSync(passedCapyNodes)

	jsonedResp, _ := json.Marshal(
		map[string]string{
			"message": "Peer UIDs synchronized successfully",
		},
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonedResp)
}
