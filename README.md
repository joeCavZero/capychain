<div align="center">
<image src="./docs/images/logo.jpg" width="500"/>
</div>

<h1 align="center">CapyChain</h1>

CapyChain is a lightweight blockchain implementation in Go, designed for educational purposes and small-scale applications. It features a simple proof-of-work algorithm, peer-to-peer networking, and a RESTful API for interaction.

## Features
- Simple Proof-of-Work consensus mechanism
- Peer-to-peer networking
- RESTful API for blockchain interaction
- Web interface for easy access and management

## Interface
The CapyChain provides a user-friendly web interface to interact with the blockchain. You can access the interface by navigating to `http://<address>:<port>/interface` in your web browser, where `<address>` and `<port>` correspond to the address and port of your CapyChain node.

## Election System
CapyChain includes a basic election system that allows nodes to vote for peers in the network. This feature is designed to demonstrate decentralized decision-making within the blockchain network.

The election model used in CapyChain is the `Voting` model, where each node can cast a vote for a peer node. The peer with the highest number of votes rules the mining difficulty for the network. 

Example:
    We have three nodes: NodeA, NodeB, and NodeC.
    - NodeA votes for NodeB
    - NodeB votes for NodeC
    - NodeC votes for NodeB
    In this scenario, NodeB receives two votes (from NodeA and NodeC), while NodeC receives one vote (from NodeB). Therefore, NodeB becomes the elected peer that sets the mining difficulty for the network.

## Synchronization
CapyChain supports synchronization of the blockchain and node information across the network. Nodes can synchronize their blockchain data and peer lists to ensure consistency and up-to-date information.

For synchronize the entire network, you need to trigger the `/sync` endpoint for each node in the network. This will ensure that all nodes have the latest blockchain data and peer information.

## Documentation
- [Installation Guide](./docs/installation.md) - Steps to install and set up CapyChain.
- [API Documentation](./docs/api.md) - Detailed information about the API endpoints available in CapyChain.