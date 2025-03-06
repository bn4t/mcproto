package types

import "fmt"

type VarInt int32

func (v VarInt) Marshal() ([]byte, error) {
	var value uint32 = uint32(v)
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

func (v *VarInt) Unmarshal(data []byte) error {
	var value uint32
	var position uint = 0

	for i := 0; i < len(data); i++ {
		currentByte := data[i]
		value |= uint32(currentByte&0x7F) << position

		if currentByte&0x80 == 0 {
			*v = VarInt(value)
			return nil
		}

		position += 7
		if position >= 32 {
			return fmt.Errorf("VarInt is too big")
		}
	}

	return fmt.Errorf("VarInt is incomplete")
}
