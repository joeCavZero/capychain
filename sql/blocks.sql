-- name: CreateBlocksTableIfNotExists :exec
CREATE TABLE IF NOT EXISTS blocks (
    height INTEGER PRIMARY KEY,
    hash TEXT NOT NULL UNIQUE,
    previous_hash TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    nonce INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    data TEXT NOT NULL
);

-- name: GetAllBlocks :many
SELECT * FROM blocks ORDER BY height DESC;

-- name: GetBlockByHash :one
SELECT * FROM blocks WHERE hash = ?;

-- name: GetBlockByHeight :one
SELECT * FROM blocks WHERE height = ?;

-- name: InsertBlock :exec
INSERT INTO blocks (
    height, 
    hash, 
    previous_hash, 
    timestamp, 
    nonce, 
    difficulty,
    data
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetBlocksLength :one
SELECT COUNT(*) as length FROM blocks;

-- name: GetBlocksOrderedByHeight :many
SELECT * FROM blocks ORDER BY height ASC;

-- name: DeleteBlockByHeightAndHash :exec
DELETE FROM blocks WHERE height = ? AND hash = ?;