# CapyChain API Documentation

**Version:** 0.1.0  
**Base URL:** `http://{HOST}:{PORT}`

---

## Quick Overview

- **Format:** JSON for all data exchange endpoints.
- **Default Header:** `Content-Type: application/json`
- **Testing:** Use any HTTP client for requests.

---

## Models (JSON Examples)

### CapyPeer
```json
{
    "address": "127.0.0.1",
    "port": "8080"
}
```

### CapyNode
```json
{
    "uid": 1,
    "name": "node-name",
    "address": "192.168.1.10",
    "port": "8080",
    "peers": [
        { "address": "192.168.1.11", "port": "8081" }
    ]
}
```

### CapyBlock
```json
{
    "height": 0,
    "hash": "0000abcd...",
    "previous_hash": "0",
    "timestamp": 1610000000,
    "nonce": 0,
    "difficulty": 0,
    "data": "Genesis Block"
}
```

---

## Endpoints

### **GET /node**
- **Description:** Returns local node information.
- **Response 200:** `CapyNode` (JSON)
- **Example Request:**
    - **Method:** GET
    - **URL:** `/node`

### **POST /peers**
- **Description:** Adds a peer to the node.
- **Example Request:**
    - **Method:** POST
    - **URL:** `/peers`
    - **Body:**
      ```json
      {
          "address": "192.168.1.11",
          "port": "8081"
      }
      ```

### **DELETE /peers**
- **Description:** Removes a peer.
- **Example Request:**
    - **Method:** DELETE
    - **URL:** `/peers`
    - **Body:**
      ```json
      {
          "address": "192.168.1.11",
          "port": "8081"
      }
      ```

### **GET /node/sync**
- **Description:** Starts peer synchronization from the local node.
- **Example Request:**
    - **Method:** GET
    - **URL:** `/node/sync`

### **POST /node/sync**
- **Description:** Receives a list of `CapyNode` to propagate UID/peer synchronization.
- **Body:** `[CapyNode]` (JSON)
- **Example Request:**
    - **Method:** POST
    - **URL:** `/node/sync`
    - **Body:**
      ```json
      [
          {
              "uid": 1,
              "name": "n",
              "address": "192.168.1.10",
              "port": "8080",
              "peers": []
          }
      ]
      ```

### **GET /chain**
- **Description:** Returns the entire local blockchain.
- **Response 200:** `[CapyBlock]`
- **Example Request:**
    - **Method:** GET
    - **URL:** `/chain`

### **POST /chain**
- **Description:** Returns blocks starting from a minimum height.
- **Body:** `{ "height": <int64> }`
- **Response 200:** `[CapyBlock]`
- **Example Request:**
    - **Method:** POST
    - **URL:** `/chain`
    - **Body:**
      ```json
      {
          "height": 10
      }
      ```

### **GET /chain/length**
- **Description:** Returns the blockchain length.
- **Response 200:** `{ "length": <int64> }`
- **Example Request:**
    - **Method:** GET
    - **URL:** `/chain/length`

### **GET /chain/sync**
- **Description:** Starts blockchain synchronization with configured peers.
- **Response 200:** No payload.
- **Example Request:**
    - **Method:** GET
    - **URL:** `/chain/sync`

### **POST /mine**
- **Description:** Mines a block with the provided data, persists it, and returns the mined block.
- **Body:** `{ "data": "<string>" }`
- **Response 200:** `CapyBlock` (mined block)
- **Example Request:**
    - **Method:** POST
    - **URL:** `/mine`
    - **Body:**
      ```json
      {
          "data": "Hello"
      }
      ```

### **POST /block**
- **Description:** Inserts the provided block directly.
- **Body:** `CapyBlock` (JSON)
- **Response 201:** Confirmation message.
- **Example Request:**
    - **Method:** POST
    - **URL:** `/block`
    - **Body:**
      ```json
      {
          "height": 1,
          "hash": "...",
          "previous_hash": "0",
          "timestamp": 1610000000,
          "nonce": 0,
          "difficulty": 0,
          "data": "..."
      }
      ```

### **DELETE /block**
- **Description:** Removes a block by height and hash.
- **Body:** `{ "height": <int64>, "hash": "<string>" }`
- **Response 200:** Confirmation message.
- **Example Request:**
    - **Method:** DELETE
    - **URL:** `/block`
    - **Body:**
      ```json
      {
          "height": 1,
          "hash": "..."
      }
      ```

### **GET /validate**
- **Description:** Validates the local blockchain.
- **Response 200:**
    - **Valid:** `{ "is_valid": true }`
    - **Invalid:** `{ "is_valid": false, "inconsistent_block": {CapyBlock} }`
- **Example Request:**
    - **Method:** GET
    - **URL:** `/validate`

### **GET /interface**
- **Description:** Simple interface that lists blocks (plain text).
- **Response 200:** Formatted text.
- **Example Request:**
    - **Method:** GET
    - **URL:** `/interface`

---

