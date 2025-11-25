-- name: CreateBlocksTableIfNotExists :exec
CREATE TABLE IF NOT EXISTS blocks (
    height INTEGER NOT NULL,
    hash TEXT NOT NULL,
    previous_hash TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    nonce INTEGER NOT NULL,
    data TEXT NOT NULL,
    UNIQUE (hash, height)
);

-- name: GetAllBlocks :many
SELECT * FROM blocks ORDER BY height DESC;

-- name: GetBlockByHash :one
SELECT * FROM blocks WHERE hash = ?;

-- name: GetBlockByHeight :one
SELECT * FROM blocks WHERE height = ?;

-- name: InsertBlock :exec
INSERT OR IGNORE INTO blocks (
    height,
    hash,
    previous_hash,
    timestamp,
    nonce,
    data
) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetBlocksLength :one
SELECT COUNT(*) as length FROM blocks;

-- name: GetAllBlocksOrderedByHeight :many
SELECT * FROM blocks ORDER BY height ASC;

-- name: DeleteBlockByHeightAndHash :exec
DELETE FROM blocks WHERE height = ? AND hash = ?;

-- name: GetBlocksWithMinHeight :many
SELECT * FROM blocks WHERE height >= ? ORDER BY height ASC;

-- name: GetHighestBlock :one
SELECT * FROM blocks ORDER BY height DESC LIMIT 1;
