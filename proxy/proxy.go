package proxy

import (
	"fmt"
	"net"
	"sync"
)

type Proxy struct {
	listener    net.Listener
	serverAddr  string
	connections sync.Map
}

// NewProxy creates a new Minecraft proxy
func NewProxy(listenAddr, serverAddr string) (*Proxy, error) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to start proxy listener: %v", err)
	}

	return &Proxy{
		listener:   listener,
		serverAddr: serverAddr,
	}, nil
}

// Start begins accepting client connections
func (p *Proxy) Start() error {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %v", err)
		}

		go p.handleConnection(conn)
	}
}

// Stop stops the proxy server
func (p *Proxy) Stop() error {
	return p.listener.Close()
}
