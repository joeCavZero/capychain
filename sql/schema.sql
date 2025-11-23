CREATE TABLE IF NOT EXISTS blocks (
    height INTEGER NOT NULL,
    hash TEXT NOT NULL,
    previous_hash TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    nonce INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    data TEXT NOT NULL
);
