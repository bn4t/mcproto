package packet

import (
	"io"
)

type Packet interface {
	ID() int32
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}

var packetRegistry = map[int32]func() Packet{}

func RegisterPacket(id int32, constructor func() Packet) {
	packetRegistry[id] = constructor
}
