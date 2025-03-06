package packet

import (
	"encoding/binary"
	"io"
	"mc-proxy/protocol/types"
)

type Handshake struct {
	ProtocolVersion types.VarInt
	ServerAddress   types.String
	ServerPort      types.UnsignedShort
	NextState       types.VarInt // 1 for status, 2 for login
}

func (p *Handshake) ID() int32 { return 0x00 }

func (p *Handshake) Encode(w io.Writer) error {

	if err := types.WriteVarInt(p.ProtocolVersion, w); err != nil {
		return err
	}

	if err := types.WriteString(p.ServerAddress, w); err != nil {
		return err
	}

	if err := types.WriteUnsignedShort(p.ServerPort, w); err != nil {
		return err
	}

	return types.WriteVarInt(p.NextState, w)
}

func (p *Handshake) Decode(r io.Reader) error {
	var err error
	if p.ProtocolVersion, err = types.ReadVarInt(r); err != nil {
		return err
	}
	if p.ServerAddress, err = ReadString(r); err != nil {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &p.ServerPort); err != nil {
		return err
	}
	p.NextState, err = ReadVarInt(r)
	return err
}

func init() {
	RegisterPacket(0x00, func() Packet { return &Handshake{} })
}
