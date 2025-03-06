package types

import (
	"encoding/binary"
	"fmt"
	"math"
)

type Double float64

func (d Double) Marshal() ([]byte, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(float64(d)))
	return buf, nil
}

func (d *Double) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("double data too short")
	}
	*d = Double(math.Float64frombits(binary.BigEndian.Uint64(data)))
	return nil
}
