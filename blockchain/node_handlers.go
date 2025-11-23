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

	// Endpoint para listar peers conectados
	r.HandleFunc(
		"/peers",
		peersGetHandler,
	).Methods("GET")

	// Endpoint para adicionar um novo peer
	r.HandleFunc(
		"/peers",
		peersPostHandler,
	).Methods("POST")
}

func nodeHandler(w http.ResponseWriter, r *http.Request) {
	/*
		Write CapyNode information as JSON to response
	*/
	jsonedNode, err := json.Marshal(CapyBlockchainInstance.Node.ToCapyNodeResponse())
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

func peersGetHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var allPeers []CapyPeer = CapyBlockchainInstance.Node.ListPeers()
	jsonedPeers, err := json.Marshal(allPeers)
	if err != nil {
		jsonedErr, _ := json.Marshal(
			map[string]string{
				"error": "Failed to marshal peers information",
			},
		)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonedErr)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonedPeers)
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
