package types

import "fmt"

type Byte int8

func (b Byte) Marshal() ([]byte, error) {
	return []byte{byte(b)}, nil
}

func (b *Byte) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("byte data too short")
	}
	*b = Byte(data[0])
	return nil
}
