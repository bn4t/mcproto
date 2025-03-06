package types

import (
	"encoding/binary"
	"fmt"
)

type Short int16

func (s Short) Marshal() ([]byte, error) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(s))
	return buf, nil
}

func (s *Short) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("short data too short")
	}
	*s = Short(binary.BigEndian.Uint16(data))
	return nil
}
