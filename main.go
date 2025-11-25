package main

import (
	"capychain/blockchain"
	"capychain/dbg"
	"os"
)

const DatabaseSource string = "./capychain.sqlite3"

func main() {
	dbg.Infof("Initializing System")

	var err error

	var nodeName string = "capychain node"
	var port string = "8080"
	if len(os.Args) >= 2 {
		nodeName = os.Args[1]
	}
	if len(os.Args) >= 3 {
		port = os.Args[2]
	}
	err = blockchain.Init(nodeName, port, DatabaseSource)
	if err != nil {
		dbg.ExitWithErrorf("Failed to start API server: %s", err.Error())
	}
}
