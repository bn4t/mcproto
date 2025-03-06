package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mc-proxy/proxy"
)

func main() {
	// Parse command line flags
	listenAddr := flag.String("listen", "127.0.0.1:25565", "Address to listen on")
	serverAddr := flag.String("server", "127.0.0.1:25566", "Address of the Minecraft server")
	flag.Parse()

	// Create and start the proxy
	p, err := proxy.NewProxy(*listenAddr, *serverAddr)
	if err != nil {
		fmt.Printf("Failed to create proxy: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start proxy in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- p.Start()
	}()

	// Wait for either an error or shutdown signal
	select {
	case err := <-errChan:
		if err != nil {
			fmt.Printf("Proxy error: %v\n", err)
			os.Exit(1)
		}
	case <-sigChan:
		fmt.Println("\nShutting down proxy...")
		if err := p.Stop(); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
			os.Exit(1)
		}
	}
}
