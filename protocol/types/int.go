package types

import (
	"encoding/binary"
	"fmt"
)

type Int int32

func (i Int) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf, nil
}

func (i *Int) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("int data too short")
	}
	*i = Int(binary.BigEndian.Uint32(data))
	return nil
}
