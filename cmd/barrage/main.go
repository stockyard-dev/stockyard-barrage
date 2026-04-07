package main

import (
	"flag"
	"fmt"
	"github.com/stockyard-dev/stockyard-barrage/internal/server"
	"github.com/stockyard-dev/stockyard-barrage/internal/store"
	"log"
	"net/http"
	"os"
)

func main() {
	portFlag := flag.String("port", "", "")
	dataFlag := flag.String("data", "", "")
	flag.Parse()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8600"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./barrage-data"
	}
	if *portFlag != "" {
		port = *portFlag
	}
	if *dataFlag != "" {
		dataDir = *dataFlag
	}
	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("barrage: %v", err)
	}
	defer db.Close()
	srv := server.New(db, server.DefaultLimits(), dataDir)
	fmt.Printf("\n  Barrage — Self-hosted load tester\n  ─────────────────────────────────\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Data:       %s\n  ─────────────────────────────────\n  Questions? hello@stockyard.dev\n\n", port, port, dataDir)
	log.Printf("barrage: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
