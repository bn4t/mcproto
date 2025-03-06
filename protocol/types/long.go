package types

import (
	"encoding/binary"
	"fmt"
)

type Long int64

func (l Long) Marshal() ([]byte, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(l))
	return buf, nil
}

func (l *Long) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("long data too short")
	}
	*l = Long(binary.BigEndian.Uint64(data))
	return nil
}
