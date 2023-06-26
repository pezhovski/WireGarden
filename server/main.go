package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wire-garden-server/internal/tunnel"
)

func main() {
	server := startServer()

	gracefulShutdown(server)
}

func startServer() *http.Server {
	serverAddress := "0.0.0.0:8080"

	err := tunnel.Bootstrap()
	if err != nil {
		log.Fatalf("Failed to bootstrap the tunnel: %v", err)
	}

	server := &http.Server{
		Addr: serverAddress,
	}

	go func() {
		log.Printf("Server is running on %s\n", serverAddress)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	return server
}

func gracefulShutdown(server *http.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	<-signals

	log.Println("Shutting down server gracefully...")

	tunnel.Teardown()
	server.Close()

	log.Println("Server stopped.")
}
