package blockchain

import (
	"capychain/dbg"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	DISCOVERY_PORT    int    = 9999
	DISCOVERY_MESSAGE string = "CAPYCHAIN_DISCOVERY"
)

type CapyNode struct {
	UID     uint64      `json:"uid"`
	Name    string      `json:"name"`
	Address string      `json:"address"`
	Port    string      `json:"port"`
	Router  *mux.Router `json:"router"`

	Peers []CapyPeer `json:"peers"`
}

func NewCapyNode(name string, port string) *CapyNode {
	return &CapyNode{
		UID:     0,
		Name:    name,
		Address: "",
		Port:    port,
		Router:  mux.NewRouter(),
		Peers:   []CapyPeer{},
	}
}

func GetLocalIP() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addresses {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no non-loopback IP address found")
}

func (cn *CapyNode) StartServer() error {
	var err error

	SetupInterfaceHandlers(cn.Router)
	SetupBlockchainHandlers(cn.Router)
	SetupNodeHandlers(cn.Router)

	cn.Address, err = GetLocalIP()
	if err != nil {
		return fmt.Errorf("failed to get local IP address: %s", err.Error())
	}

	dbg.Infof("Starting node server on %s:%s", cn.Address, cn.Port)
	err = http.ListenAndServe(
		fmt.Sprintf(":%s", cn.Port),
		cn.Router,
	)
	if err != nil {
		return err
	}

	return nil
}

func (cn *CapyNode) AddCapyPeer(peer CapyPeer) {
	for _, p := range cn.Peers {
		if p.Address == peer.Address && p.Port == peer.Port {
			return
		}
	}
	cn.Peers = append(cn.Peers, peer)
}

func (cn *CapyNode) ListPeers() []CapyPeer {
	return cn.Peers
}

type CapyPeer struct {
	Address string `json:"address"`
	Port    string `json:"port"`
}

func NewCapyPeer(address string, port string) CapyPeer {
	return CapyPeer{
		Address: address,
		Port:    port,
	}
}

func (cp *CapyPeer) FetchPeers() ([]CapyPeer, error) {
	var err error

	var resp *http.Response
	resp, err = http.Get(
		fmt.Sprintf("http://%s:%s/peers", cp.Address, cp.Port),
	)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("error fetching peers from peer [%s:%s]: %s", cp.Address, cp.Port, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("non-OK HTTP status from peer [%s:%s]: %s", cp.Address, cp.Port, resp.Status)
	}

	var peerPeers []CapyPeer
	err = json.NewDecoder(resp.Body).Decode(&peerPeers)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("error decoding peers from peer [%s:%s]: %s", cp.Address, cp.Port, err.Error())
	}
	resp.Body.Close()

	return peerPeers, nil
}

type CapyNodeResponse struct {
	UID     uint64 `json:"uid"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    string `json:"port"`
}

func NewCapyNodeResponse(uid uint64, name string, address string, port string) *CapyNodeResponse {
	return &CapyNodeResponse{
		UID:     uid,
		Name:    name,
		Address: address,
		Port:    port,
	}
}

func (node *CapyNode) ToCapyNodeResponse() *CapyNodeResponse {
	return NewCapyNodeResponse(
		node.UID,
		node.Name,
		node.Address,
		node.Port,
	)
}

func (cn *CapyNode) SynchronizePeers() error {
	var err error
	for _, peer := range cn.Peers {
		dbg.Infof("Synchronizing with peer %s:%s", peer.Address, peer.Port)
		var peerPeers []CapyPeer
		peerPeers, err = peer.FetchPeers()
		if err != nil {
			return fmt.Errorf("error synchronizing peers from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
		}
		for _, p := range peerPeers {
			cn.AddCapyPeer(p)
		}
	}
	return nil
}

func (cn *CapyNode) SynchronizeUIDs(passedUIDs []uint64) error {
	return nil
}
