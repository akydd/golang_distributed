package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"server/node"
	"syscall"
	"time"
)

var (
	nodeID   string
	bindAddr string
	raftAddr string
	joinAddr string
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from %s\n", r.URL.Path)
}

type Join struct {
	Address string `json:"address"`
	NodeID  string `json:"node_id"`
}

func main() {
	flag.StringVar(&nodeID, "id", "", "unique node identifier")
	flag.StringVar(&bindAddr, "baddr", "", "HTTP bind address for the node")
	flag.StringVar(&raftAddr, "raddr", "", "raft bind address for the node")
	flag.StringVar(&joinAddr, "jaddr", "", "cluster address to join")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	node, err := node.New(&node.Config{
		ID:       nodeID,
		RaftAddr: raftAddr,
		JoinAddr: joinAddr,
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", testHandler)
	mux.HandleFunc("POST /join", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var j Join

		err := json.NewDecoder(r.Body).Decode(&j)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err = node.Join(j.Address, j.NodeID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    bindAddr,
		Handler: mux,
	}

	go func() {
		fmt.Printf("Server started at address %s\n", bindAddr)
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	if joinAddr != "" {
		postBody, _ := json.Marshal(map[string]string{
			"address": joinAddr,
			"node_id": nodeID,
		})
		b := bytes.NewBuffer(postBody)
		_, err := http.Post("http://"+joinAddr+"/join", "application-type/json", b)
		if err != nil {
			log.Fatal(err)
		}
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Forced shutdown ...")
	} else {
		fmt.Println("Graceful shutdown ...")
	}
}
