package blockchain

import (
	"capychain/dbg"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	DISCOVERY_PORT    int    = 9999
	DISCOVERY_MESSAGE string = "CAPYCHAIN_DISCOVERY"
)

type CapyNode struct {
	UID    uint64      `json:"uid"`
	Name   string      `json:"name"`
	Port   string      `json:"port"`
	Router *mux.Router `json:"router"`

	Peers []CapyPeer `json:"peers"`
}

func NewCapyNode(name string, port string) *CapyNode {
	return &CapyNode{
		UID:    0,
		Name:   name,
		Port:   port,
		Router: mux.NewRouter(),
		Peers:  []CapyPeer{},
	}
}

func (cn *CapyNode) StartServer() error {
	var err error

	SetupInterfaceHandlers(cn.Router)
	SetupBlockchainHandlers(cn.Router)
	SetupNodeHandlers(cn.Router)

	dbg.Infof("Starting node server on port %s", cn.Port)
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

func (cn *CapyNode) StartDiscoveryListener() {
	addr := net.UDPAddr{
		Port: DISCOVERY_PORT,
		IP:   net.IPv4zero,
	}

	conn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		dbg.Errorf("UDP discovery listener error: %s", err)
		return
	}
	dbg.Infof("Discovery listener running on UDP port %d", DISCOVERY_PORT)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			msg := string(buf[:n])
			if msg == DISCOVERY_MESSAGE {
				// Responde com IP e porta HTTP do nó atual
				response := fmt.Sprintf("NODE:%s:%s", cn.Name, cn.Port)
				conn.WriteToUDP([]byte(response), remoteAddr)
			}
		}
	}()
}

func (cn *CapyNode) ScanNetworkForPeers() error {
	addr := net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: DISCOVERY_PORT,
	}

	conn, err := net.DialUDP("udp4", nil, &addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 1. Envia broadcast
	_, err = conn.Write([]byte(DISCOVERY_MESSAGE))
	if err != nil {
		return err
	}

	// 2. Configura timeout para ouvir respostas
	err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)

	for {
		// 3. Espera respostas
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			break // timeout → encerra
		}

		msg := string(buf[:n])

		if strings.HasPrefix(msg, "NODE:") {
			parts := strings.Split(msg, ":")
			if len(parts) != 3 {
				continue
			}

			name := parts[1]
			port := parts[2]

			newPeer := CapyPeer{
				Address: remoteAddr.IP.String(),
				Port:    port,
			}

			cn.AddCapyPeer(newPeer)
			dbg.Infof("Discovered peer %s (%s:%s)", name, newPeer.Address, newPeer.Port)
		}
	}

	return nil
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
