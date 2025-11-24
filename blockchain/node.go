package blockchain

import (
	"bytes"
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
	Router  *mux.Router `json:"-"`

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

// A função abaixo ajeita os UIDs dos nodes / peers da rede
func (cn *CapyNode) SyncNodePeers() {
	var passedCapyNodes []CapyNode = []CapyNode{
		*cn,
	}

	for _, peer := range cn.Peers {
		peer.CastNodePeersSync(passedCapyNodes)
	}
}

func GetNextUID(capyNodes []CapyNode) uint64 {
	var maxUID uint64 = 0
	for _, cpyNd := range capyNodes {
		if cpyNd.UID > maxUID {
			maxUID = cpyNd.UID
		}
	}
	return maxUID + 1
}

func (cn *CapyNode) CastNodePeersSync(passedCapyNodes []CapyNode) {
	for _, cpyNd := range passedCapyNodes {
		if cpyNd.Address == cn.Address && cpyNd.Port == cn.Port {
			return
		}

		if cpyNd.UID == cn.UID {
			cn.UID = GetNextUID(passedCapyNodes)
		}

		newCapyPeer := NewCapyPeer(cpyNd.Address, cpyNd.Port)
		cn.AddCapyPeer(newCapyPeer)

	}

	passedCapyNodes = append(passedCapyNodes, *cn)

	for _, peer := range cn.Peers {
		peer.CastNodePeersSync(passedCapyNodes)
	}
}

func (cp *CapyPeer) CastNodePeersSync(passedCapyNodes []CapyNode) {
	var err error
	jsonedPassedCapyNodes, err := json.Marshal(passedCapyNodes)
	if err != nil {
		return
	}
	dbg.Infof("Casting peer sync to %s:%s", cp.Address, cp.Port)
	_, err = http.Post(
		fmt.Sprintf(
			"http://%s:%s/node/sync",
			cp.Address, cp.Port,
		),
		"application/json",
		bytes.NewBuffer(jsonedPassedCapyNodes),
	)
	if err != nil {
		return
	}
}
