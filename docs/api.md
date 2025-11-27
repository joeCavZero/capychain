# CapyChain API Documentation

This document provides an overview of the API endpoints available in the CapyChain blockchain application. The API allows interaction with the blockchain, including retrieving blocks, adding transactions, and managing nodes.

## Most Used Request and Response Models
### CapyBlock
Represents a block in the blockchain.
```json
{
  "height": <integer>,
  "hash": "<string>",
  "previous_hash": "<string>",
  "timestamp": "<string>",
  "nonce": <integer>,
  "data": "<string>"  
}
```

### CapyNode
Represents a node in the blockchain network.
```json
{
    "name": "<string>",
    "address": "<string>",
    "port": "<string>",
    "peers": [<CapyPeer>, ...],
    "vote": <CapyPeer>,
    "difficulty": <integer>
}
```

### CapyPeer
Represents a peer node in the blockchain network.
```json
{
    "address": "<string>",
    "port": "<string>"
}
```

## Endpoints
- `GET /interface`: Retrieve the API interface.
    - Response: Web page
- `GET /chain`: Retrieve the entire blockchain.
    - Response: 
        - Status 200: 
            - Body: JSON array of `CapyBlock`
        - Status 500: Internal server error
- `POST /chain`: Get a block by its height and hash.
    - Request Body:
      ```json
      {
        "height": <integer>,
        "hash": "<block_hash>"
      }
      ```
    - Response: 
        - Status 200:
            - Body: `CapyBlock`
        - Status 400: Bad request (invalid input)
        - Status 500: Internal server error
- `GET /chain/length`: Get the length of the blockchain.
    - Response: 
        - Status 200:
            - Body:
                ```json
                {
                    "length": <integer>
                }
                ```
        - Status 500: Internal server error
- `POST /chain/block`: Add a new block to the blockchain.
    - Request Body: `CapyBlock`
    - Response:
        - Status 201: Block successfully created
        - Status 400: Bad request (invalid block data)
        - Status 500: Internal server error
- `GET /chain/sync`: Synchronize the blockchain with other nodes.
    - Response:
        - Status 200: Synchronization successful
- `POST /chain/mine`: Start the mining process for a new block.
    - Request Body:
      ```json
      {
        "data": "<string>"
      }
      ```
    - Response:
        - Status 200:
            - Body: `CapyBlock`
        - Status 400: Bad request (invalid input)
        - Status 500: Internal server error
- `DELETE /chain/block`: Delete a specific block from the blockchain.
    - Request Body:
      ```json
      {
        "height": <integer>,
        "hash": "<string>"
      }
      ```
    - Response:
        - Status 200: Block successfully deleted
        - Status 400: Bad request (invalid input)
        - Status 500: Internal server error
- `GET /chain/validate`: Validate the blockchain.
    - Response:
        - Status 200:
            - Body:
                ```json
                {
                    "is_valid": <boolean>,
                    "inconsistent_block": <CapyBlock> // Optional, only if invalid
                }
                ```
        - Status 500: Internal server error
- `GET /sync`: Synchronize all nodes.
    - Response:
        - Status 200: Synchronization successful
- `GET /node`: Retrieve information about the current node.
    - Response:
        - Status 200:
            - Body: `CapyNode`
        - Status 500: Internal server error
- `POST /node/peers`: Add a new peer to the node.
    - Request Body:
      ```json
      {
        "address": "<string>",
        "port": "<string>"
      }
      ```
    - Response:
        - Status 200: Peer successfully added
        - Status 400: Bad request (invalid input)
- `DELETE /node/peers`: Remove a peer from the node.
    - Request Body:
      ```json
      {
        "address": "<string>",
        "port": "<string>"
      }
      ```
    - Response:
        - Status 200: Peer successfully removed
        - Status 400: Bad request (invalid input)
- `GET /node/sync`: Start the synchronization process for the node.
    - Response:
        - Status 200: Synchronization successful
- `POST /node/sync`: Synchronize the node with a list of other nodes.
    - Request Body:
      ```json
      [
        <CapyNode>, ...
      ]
      ```
    - Response:
        - Status 200: Synchronization successful
        - Status 400: Bad request (invalid input)
- `POST /node/vote`: Vote for a peer node.
    - Request Body:
      ```json
      {
        "address": "<string>",
        "port": "<string>"
      }
      ```
    - Response:
        - Status 200: Vote successfully cast
        - Status 400: Bad request (invalid input)
- `POST /node/difficulty`: Set the mining difficulty for the node.
    - Request Body:
      ```json
      {
        "difficulty": <integer>
      }
      ```
    - Response:
        - Status 200: Difficulty successfully updated
        - Status 400: Bad request (invalid input)
