package blockchain

import (
	"capychain/dbg"
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

	// Endpoint para remover um peer
	r.HandleFunc(
		"/peers",
		peersDeleteHandler,
	).Methods("DELETE")

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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dbg.Infof("Providing node information")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonedNode)
}

func peersPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var newPeer CapyPeer
	err = json.NewDecoder(r.Body).Decode(&newPeer)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbg.Infof("Adding new peer [%s:%s]", newPeer.Address, newPeer.Port)
	CapyBlockchainInstance.Node.AddCapyPeer(newPeer)
	w.WriteHeader(http.StatusOK)
}

func nodeSyncGetHandler(w http.ResponseWriter, r *http.Request) {
	dbg.Infof("Starting node synchronization process")
	CapyBlockchainInstance.Node.SyncNodePeers()
	w.WriteHeader(http.StatusOK)
}

func nodeSyncPostHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var passedCapyNodes []CapyNode
	err = json.NewDecoder(r.Body).Decode(&passedCapyNodes)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbg.Infof("Synchronizing nodes")
	CapyBlockchainInstance.Node.CastNodePeersSync(passedCapyNodes)
	w.WriteHeader(http.StatusOK)
}

func peersDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var peerToRemove CapyPeer
	err = json.NewDecoder(r.Body).Decode(&peerToRemove)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbg.Infof("Removing peer [%s:%s]", peerToRemove.Address, peerToRemove.Port)
	CapyBlockchainInstance.Node.RemoveCapyPeer(peerToRemove)
	w.WriteHeader(http.StatusOK)
}
