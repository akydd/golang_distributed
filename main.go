package main

import (
	"context"
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
	port     int
	nodeID   string
	bindAddr string
	raftAddr string
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from %s\n", r.URL.Path)
}

func main() {
	flag.IntVar(&port, "port", 8080, "port for this service")
	flag.StringVar(&nodeID, "id", "", "unique node identifier")
	flag.StringVar(&bindAddr, "baddr", "", "HTTP bind address for the node")
	flag.StringVar(&raftAddr, "raddr", "", "raft bind address for the node")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", testHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	_, err := node.New(&node.Config{
		ID:       nodeID,
		BindAddr: bindAddr,
		RaftAddr: raftAddr,
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		fmt.Printf("Server started on port %d\n", port)
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Forced shutdown ...")
	} else {
		fmt.Println("Graceful shutdown ...")
	}
}
