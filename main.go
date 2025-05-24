package main

import (
	"log"
	"net"
	"net/http"

	"github.com/FelipeMCassiano/RedicaMq/config"
	"github.com/FelipeMCassiano/RedicaMq/queue"
)

func main() {
	l, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	log.Printf("listening on ws://%v", l.Addr())
	s := &http.Server{
		Handler: queue.NewQueueServer(config.LoadConfig()),
	}

	errc := make(chan error, 1)

	go func() {
		errc <- s.Serve(l)
	}()

	if err := <-errc; err != nil {
		log.Fatalf("server error: %v\n", err)
	}
}
