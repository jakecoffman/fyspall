package main

import (
	"net/http"
	"log"
	"github.com/jakecoffman/fyspall"
	"net/http/pprof"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	server := spyfall.NewServer()
	addr := "0.0.0.0:3032"
	log.Println("Serving on", addr)
	server.HandleFunc("/debug/pprof/", pprof.Index)
	server.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	server.HandleFunc("/debug/pprof/profile", pprof.Profile)
	server.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	log.Fatal(http.ListenAndServe(addr, server))
}
