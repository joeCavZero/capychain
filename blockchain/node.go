package blockchain

import (
	"bytes"
	"capychain/dbg"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

const (
	DISCOVERY_PORT    int    = 9999
	DISCOVERY_MESSAGE string = "CAPYCHAIN_DISCOVERY"
)

type CapyNode struct {
	UID     uint64 `json:"uid"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    string `json:"port"`

	Peers      []CapyPeer `json:"peers"`
	Vote       CapyPeer   `json:"vote"`
	Difficulty int        `json:"difficulty"`
}

func NewCapyNode(name string, port string) *CapyNode {
	localAddress, err := GetLocalAddress()
	if err != nil {
		localAddress = ""
	}
	votePeer := NewCapyPeer(localAddress, port)
	return &CapyNode{
		UID:     0,
		Name:    name,
		Address: localAddress,
		Port:    port,

		Peers:      []CapyPeer{},
		Vote:       votePeer,
		Difficulty: 1,
	}

}

func GetLocalAddress() (string, error) {
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

func (cn *CapyNode) RemoveCapyPeer(peerToRemove CapyPeer) {
	var updatedPeers []CapyPeer
	for _, peer := range cn.Peers {
		if peer.Address == peerToRemove.Address && peer.Port == peerToRemove.Port {
			continue
		}
		updatedPeers = append(updatedPeers, peer)
	}
	cn.Peers = updatedPeers
}

func (cn *CapyNode) VoteForPeer(peerVote CapyPeer) {
	cn.Vote = peerVote
}

func (cn *CapyNode) GetMostVotedNodeMiningDifficulty() int {
	cn.SyncNodePeers()

	var votes map[CapyPeer]uint64 = make(map[CapyPeer]uint64)

	peers := cn.Peers
	peers = append(
		peers,
		NewCapyPeer(cn.Address, cn.Port),
	)
	for _, peer := range peers {
		dbg.Infof("Fetching vote from peer [%s:%s]", peer.Address, peer.Port)
		peerNodeResp, err := http.Get(
			fmt.Sprintf("http://%s:%s/node", peer.Address, peer.Port),
		)
		if err != nil {
			dbg.Errorf("Error fetching node info from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
			continue
		}
		if peerNodeResp.StatusCode != http.StatusOK {
			dbg.Errorf("Non-OK HTTP status from peer [%s:%s]: %s", peer.Address, peer.Port, peerNodeResp.Status)
			peerNodeResp.Body.Close()
			continue
		}

		var peerNode CapyNode
		err = json.NewDecoder(peerNodeResp.Body).Decode(&peerNode)
		if err != nil {
			dbg.Errorf("Error decoding node info from peer [%s:%s]: %s", peer.Address, peer.Port, err.Error())
			peerNodeResp.Body.Close()
			continue
		}
		peerNodeResp.Body.Close()

		v, ok := votes[peerNode.Vote]
		if ok {
			votes[peerNode.Vote] = v + 1
		} else {
			votes[peerNode.Vote] = 1
		}
	}

	var selectedPeer CapyPeer
	var maxVotes uint64 = 0
	for p, v := range votes {
		if v > maxVotes {
			maxVotes = v
			selectedPeer = p
		}
	}

	// isso significa que ninguem votou em ninguem, logo, dificuldade padrao
	if maxVotes == 0 {
		return 1
	}

	dbg.Infof("Selected peer [%s:%s] with %d votes for mining difficulty", selectedPeer.Address, selectedPeer.Port, maxVotes)

	peerNodeResp, err := http.Get(
		fmt.Sprintf("http://%s:%s/node", selectedPeer.Address, selectedPeer.Port),
	)

	if err != nil {
		dbg.Errorf("Error fetching node info from selected peer [%s:%s]: %s", selectedPeer.Address, selectedPeer.Port, err.Error())
		return 1
	}
	if peerNodeResp.StatusCode != http.StatusOK {
		dbg.Errorf("Non-OK HTTP status from selected peer [%s:%s]: %s", selectedPeer.Address, selectedPeer.Port, peerNodeResp.Status)
		peerNodeResp.Body.Close()
		return 1
	}

	var selectedPeerNode CapyNode
	err = json.NewDecoder(peerNodeResp.Body).Decode(&selectedPeerNode)
	if err != nil {
		dbg.Errorf("Error decoding node info from selected peer [%s:%s]: %s", selectedPeer.Address, selectedPeer.Port, err.Error())
		peerNodeResp.Body.Close()
		return 1
	}
	peerNodeResp.Body.Close()

	return selectedPeerNode.Difficulty
}
