package types

import (
	"fmt"
	"io"
)

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

func WriteVarInt(varInt VarInt, w io.Writer) error {
	buf, err := varInt.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

func ReadVarInt(r io.Reader) (VarInt, error) {
	var value VarInt
	var buf []byte
	for {
		tmp := make([]byte, 1)
		_, err := r.Read(tmp)
		if err != nil {
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
