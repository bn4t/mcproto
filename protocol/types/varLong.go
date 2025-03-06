package types

import (
	"fmt"
)

type VarLong int64

func (v VarLong) Marshal() ([]byte, error) {
	var value uint64 = uint64((v << 1) ^ (v >> 63)) // zigzag encoding
	var buf []byte
	for {
		b := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if value == 0 {
			break
		}
	}
	return buf, nil
}

func (v *VarLong) Unmarshal(data []byte) error {
	var value uint64
	var position uint = 0

	for i := 0; i < len(data); i++ {
		currentByte := data[i]
		value |= uint64(currentByte&0x7F) << position

		if currentByte&0x80 == 0 {
			// Convert back from zigzag encoding
			*v = VarLong((value >> 1) ^ -(value & 1))
			return nil
		}

		position += 7
		if position >= 64 {
			return fmt.Errorf("VarLong is too big")
		}
	}

	return fmt.Errorf("VarLong is incomplete")
}
