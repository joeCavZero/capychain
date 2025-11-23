package blockchain

import (
	"capychain/db"
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type CapyDatabase struct {
	Database *sql.DB
}

func NewCapyDatabase(dbSource string) (*CapyDatabase, error) {
	var err error

	db, err := sql.Open("sqlite3", dbSource)
	if err != nil {
		return nil, err
	}

	cd := &CapyDatabase{
		Database: db,
	}

	dt := cd.NewQueries()
	ctx := context.Background()
	err = dt.CreateBlocksTableIfNotExists(ctx)
	if err != nil {
		return nil, err
	}

	return cd, nil
}

func (cd *CapyDatabase) NewQueries() *db.Queries {
	return db.New(cd.Database)
}
