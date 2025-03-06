package types

import (
	"encoding/binary"
	"fmt"
)

type UUID struct {
	MostSignificantBits  int64
	LeastSignificantBits int64
}

func (u UUID) Marshal() ([]byte, error) {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[0:8], uint64(u.MostSignificantBits))
	binary.BigEndian.PutUint64(buf[8:16], uint64(u.LeastSignificantBits))
	return buf, nil
}

func (u *UUID) Unmarshal(data []byte) error {
	if len(data) < 16 {
		return fmt.Errorf("UUID data too short")
	}

	u.MostSignificantBits = int64(binary.BigEndian.Uint64(data[0:8]))
	u.LeastSignificantBits = int64(binary.BigEndian.Uint64(data[8:16]))
	return nil
}

// String returns the string representation of the UUID
func (u UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uint32(u.MostSignificantBits>>32),
		uint16(u.MostSignificantBits>>16),
		uint16(u.MostSignificantBits),
		uint16(u.LeastSignificantBits>>48),
		uint64(u.LeastSignificantBits)&0x0000FFFFFFFFFFFF)
}
