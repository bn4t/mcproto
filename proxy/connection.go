package proxy

import (
	"fmt"
	"io"
	"net"
	"sync"

	"mc-proxy/protocol/types"
)

type clientState int

const (
	stateHandshake clientState = iota
	stateStatus
	stateLogin
	statePlay
)

type Connection struct {
	clientConn net.Conn
	serverConn net.Conn
	state      clientState
	proxy      *Proxy
	closed     bool
	mutex      sync.Mutex
}

func (p *Proxy) handleConnection(clientConn net.Conn) {
	conn := &Connection{
		clientConn: clientConn,
		proxy:      p,
		state:      stateHandshake,
	}

	defer conn.close()

	// Handle initial handshake
	if err := conn.handleHandshake(); err != nil {
		fmt.Printf("Handshake error: %v\n", err)
		return
	}

	// Based on the state after handshake, handle accordingly
	switch conn.state {
	case stateStatus:
		if err := conn.handleStatus(); err != nil {
			fmt.Printf("Status error: %v\n", err)
		}
	case stateLogin:
		if err := conn.handleLogin(); err != nil {
			fmt.Printf("Login error: %v\n", err)
		}
	}
}

func (c *Connection) readVarInt() (types.VarInt, error) {
	var value types.VarInt
	buf := make([]byte, 0, 5)
	tmp := make([]byte, 1)

	for i := 0; i < 5; i++ {
		n, err := c.clientConn.Read(tmp)
		if err != nil || n == 0 {
			return 0, fmt.Errorf("failed to read VarInt: %v", err)
		}

		buf = append(buf, tmp[0])
		if tmp[0]&0x80 == 0 {
			break
		}
	}

	if err := value.Unmarshal(buf); err != nil {
		return 0, err
	}
	return value, nil
}

func (c *Connection) readPacket() ([]byte, error) {
	// Read packet length
	length, err := c.readVarInt()
	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %v", err)
	}

	// Read the entire packet
	packet := make([]byte, length)
	_, err = io.ReadFull(c.clientConn, packet)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet: %v", err)
	}

	return packet, nil
}

func (c *Connection) handleHandshake() error {
	packet, err := c.readPacket()
	if err != nil {
		return fmt.Errorf("failed to read handshake packet: %v", err)
	}

	var pos int

	// Read packet ID
	var packetID types.VarInt
	if err := packetID.Unmarshal(packet[pos:]); err != nil {
		return fmt.Errorf("failed to read packet ID: %v", err)
	}
	// Increment position by counting VarInt bytes
	for i := pos; i < len(packet); i++ {
		pos++
		if packet[i]&0x80 == 0 {
			break
		}
	}

	// Read protocol version
	var protocolVersion types.VarInt
	if err := protocolVersion.Unmarshal(packet[pos:]); err != nil {
		return fmt.Errorf("failed to read protocol version: %v", err)
	}
	// Increment position by counting VarInt bytes
	for i := pos; i < len(packet); i++ {
		pos++
		if packet[i]&0x80 == 0 {
			break
		}
	}

	// Read server address
	var serverAddr types.String
	if err := serverAddr.Unmarshal(packet[pos:]); err != nil {
		return fmt.Errorf("failed to read server address: %v", err)
	}
	// String length is prefixed by a VarInt
	strLenStart := pos
	for i := pos; i < len(packet); i++ {
		pos++
		if packet[i]&0x80 == 0 {
			break
		}
	}
	var strLen types.VarInt
	strLen.Unmarshal(packet[strLenStart:pos])
	pos += int(strLen) // Skip the string content

	// Read server port
	var serverPort types.UnsignedShort
	if err := serverPort.Unmarshal(packet[pos:]); err != nil {
		return fmt.Errorf("failed to read server port: %v", err)
	}
	pos += 2 // UnsignedShort is always 2 bytes

	// Read next state
	var nextState types.VarInt
	if err := nextState.Unmarshal(packet[pos:]); err != nil {
		return fmt.Errorf("failed to read next state: %v", err)
	}

	c.state = clientState(nextState)

	// If we're going to login state, connect to the actual server
	if c.state == stateLogin {
		serverConn, err := net.Dial("tcp", c.proxy.serverAddr)
		if err != nil {
			return fmt.Errorf("failed to connect to server: %v", err)
		}
		c.serverConn = serverConn

		// Forward the original packet with length prefix
		lengthBytes, err := types.VarInt(len(packet)).Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal packet length: %v", err)
		}

		if _, err := c.serverConn.Write(lengthBytes); err != nil {
			return fmt.Errorf("failed to write packet length: %v", err)
		}
		if _, err := c.serverConn.Write(packet); err != nil {
			return fmt.Errorf("failed to write packet: %v", err)
		}
	}

	return nil
}

func (c *Connection) handleStatus() error {
	// Connect to the server for status requests
	serverConn, err := net.Dial("tcp", c.proxy.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	c.serverConn = serverConn
	defer c.serverConn.Close()

	// Forward handshake packet
	packet, err := c.readPacket()
	if err != nil {
		return fmt.Errorf("failed to read status packet: %v", err)
	}

	// Forward the packet with length prefix
	lengthBytes, err := types.VarInt(len(packet)).Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal packet length: %v", err)
	}

	if _, err := c.serverConn.Write(lengthBytes); err != nil {
		return fmt.Errorf("failed to write packet length: %v", err)
	}
	if _, err := c.serverConn.Write(packet); err != nil {
		return fmt.Errorf("failed to write packet: %v", err)
	}

	// Forward remaining packets in both directions
	errChan := make(chan error, 2)

	// Client -> Server
	go func() {
		for {
			packet, err := c.readPacket()
			if err != nil {
				errChan <- err
				return
			}

			lengthBytes, err := types.VarInt(len(packet)).Marshal()
			if err != nil {
				errChan <- err
				return
			}

			if _, err := c.serverConn.Write(lengthBytes); err != nil {
				errChan <- err
				return
			}
			if _, err := c.serverConn.Write(packet); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Server -> Client
	go func() {
		for {
			// Read packet length
			length, err := c.readVarIntFrom(c.serverConn)
			if err != nil {
				errChan <- err
				return
			}

			// Read packet data
			packet := make([]byte, length)
			if _, err := io.ReadFull(c.serverConn, packet); err != nil {
				errChan <- err
				return
			}

			// Forward length and packet
			lengthBytes, err := types.VarInt(len(packet)).Marshal()
			if err != nil {
				errChan <- err
				return
			}

			if _, err := c.clientConn.Write(lengthBytes); err != nil {
				errChan <- err
				return
			}
			if _, err := c.clientConn.Write(packet); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Wait for any error
	err = <-errChan
	if err == io.EOF {
		return nil
	}
	return err
}

func (c *Connection) readVarIntFrom(conn net.Conn) (types.VarInt, error) {
	var value types.VarInt
	buf := make([]byte, 0, 5)
	tmp := make([]byte, 1)

	for i := 0; i < 5; i++ {
		n, err := conn.Read(tmp)
		if err != nil || n == 0 {
			return 0, fmt.Errorf("failed to read VarInt: %v", err)
		}

		buf = append(buf, tmp[0])
		if tmp[0]&0x80 == 0 {
			break
		}
	}

	if err := value.Unmarshal(buf); err != nil {
		return 0, err
	}
	return value, nil
}

func (c *Connection) handleLogin() error {
	// Start forwarding packets in both directions
	errChan := make(chan error, 2)

	// Client -> Server
	go func() {
		for {
			packet, err := c.readPacket()
			if err != nil {
				errChan <- err
				return
			}

			lengthBytes, err := types.VarInt(len(packet)).Marshal()
			if err != nil {
				errChan <- err
				return
			}

			if _, err := c.serverConn.Write(lengthBytes); err != nil {
				errChan <- err
				return
			}
			if _, err := c.serverConn.Write(packet); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Server -> Client
	go func() {
		for {
			// Read packet length
			length, err := c.readVarIntFrom(c.serverConn)
			if err != nil {
				errChan <- err
				return
			}

			// Read packet data
			packet := make([]byte, length)
			if _, err := io.ReadFull(c.serverConn, packet); err != nil {
				errChan <- err
				return
			}

			// Forward length and packet
			lengthBytes, err := types.VarInt(len(packet)).Marshal()
			if err != nil {
				errChan <- err
				return
			}

			if _, err := c.clientConn.Write(lengthBytes); err != nil {
				errChan <- err
				return
			}
			if _, err := c.clientConn.Write(packet); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Wait for any error
	err := <-errChan
	if err == io.EOF {
		return nil
	}
	return err
}

func (c *Connection) close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.closed {
		if c.clientConn != nil {
			c.clientConn.Close()
		}
		if c.serverConn != nil {
			c.serverConn.Close()
		}
		c.closed = true
	}
}
