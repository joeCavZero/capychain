package blockchain

import (
	"capychain/dbg"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

func SetupInterfaceHandlers(r *mux.Router) {
	r.HandleFunc(
		"/interface",
		interfaceHandler,
	).Methods("GET")
}

func interfaceHandler(w http.ResponseWriter, r *http.Request) {
	dbg.Infof("Serving interface page to %s", r.RemoteAddr)
	rawSrc, err := os.ReadFile("interface.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error loading interface")
		return
	}

	src := string(rawSrc)
	src = strings.ReplaceAll(src, "%%PORT%%", CapyBlockchainInstance.Node.Port)
	src = strings.ReplaceAll(src, "%%ADDRESS%%", CapyBlockchainInstance.Node.Address)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, src)
}
