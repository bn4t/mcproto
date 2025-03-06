package types

import (
	"fmt"
)

type Position struct {
	X int32
	Y int32
	Z int32
}

func (p Position) Marshal() ([]byte, error) {
	val := uint64((uint64(p.X&0x3FFFFFF) << 38) | (uint64(p.Z&0x3FFFFFF) << 12) | uint64(p.Y&0xFFF))
	return []byte{
		byte(val >> 56),
		byte(val >> 48),
		byte(val >> 40),
		byte(val >> 32),
		byte(val >> 24),
		byte(val >> 16),
		byte(val >> 8),
		byte(val),
	}, nil
}

func (p *Position) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("position data too short")
	}

	val := uint64(data[0])<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 | uint64(data[3])<<32 |
		uint64(data[4])<<24 | uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])

	x := int32(val >> 38)
	y := int32(val & 0xFFF)
	z := int32(val << 26 >> 38)

	// Sign extension for negative values
	if x >= 1<<25 {
		x -= 1 << 26
	}
	if y >= 1<<11 {
		y -= 1 << 12
	}
	if z >= 1<<25 {
		z -= 1 << 26
	}

	p.X = x
	p.Y = y
	p.Z = z

	return nil
}
