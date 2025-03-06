package types

import (
	"encoding/binary"
	"fmt"
	"math"
)

type Float float32

func (f Float) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, math.Float32bits(float32(f)))
	return buf, nil
}

func (f *Float) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("float data too short")
	}
	*f = Float(math.Float32frombits(binary.BigEndian.Uint32(data)))
	return nil
}
