package protocol

import "mc-proxy/protocol/types"

// Packet is an uncompressed mc packet
type Packet struct {
	Length types.VarInt
	ID     types.VarInt
	Data   []byte
}
