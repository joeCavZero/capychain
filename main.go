package main

import (
	"capychain/blockchain"
	"capychain/dbg"
	"os"
)

const API_PORT string = "8080"
const DATABASE_SOURCE string = "./capychain.sqlite3"

func main() {
	dbg.Infof("Initializing System")

	var err error

	var nodeName string = ""
	if len(os.Args) >= 2 {
		nodeName = os.Args[1]
	}
	err = blockchain.Init(nodeName, API_PORT, DATABASE_SOURCE)
	if err != nil {
		dbg.ExitWithErrorf("Failed to start API server: %s", err.Error())
	}
}
