package types

import (
	"encoding/binary"
	"fmt"
)

type UnsignedShort uint16

func (us UnsignedShort) Marshal() ([]byte, error) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(us))
	return buf, nil
}

func (us *UnsignedShort) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("unsigned short data too short")
	}
	*us = UnsignedShort(binary.BigEndian.Uint16(data))
	return nil
}
