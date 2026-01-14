package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"server/node"
)

var (
	port     int
	nodeID   string
	bindAddr string
)

func main() {
	flag.IntVar(&port, "port", 8080, "port for this service")
	flag.StringVar(&nodeID, "id", "", "unique node identifier")
	flag.StringVar(&bindAddr, "baddr", "", "bind address for the node")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s", r.URL.Path)
	})

	_, err := node.New(&node.Config{
		ID:       nodeID,
		BindAddr: bindAddr,
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server started on port %d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
