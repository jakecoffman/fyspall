package main

import (
	"net/http"
	"log"
	"github.com/jakecoffman/spyfall"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	server := spyfall.NewServer()
	addr := "0.0.0.0:3032"
	log.Println("Serving on", addr)
	log.Fatal(http.ListenAndServe(addr, server))
}
