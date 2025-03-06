package types

import "fmt"

type UnsignedByte uint8

func (ub UnsignedByte) Marshal() ([]byte, error) {
	return []byte{byte(ub)}, nil
}

func (ub *UnsignedByte) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("unsigned byte data too short")
	}
	*ub = UnsignedByte(data[0])
	return nil
}
