package types

import "fmt"

type String struct {
	Value string
}

func (s String) Marshal() ([]byte, error) {
	strBytes := []byte(s.Value)
	length := VarInt(len(strBytes))

	// Marshal the length
	lenBytes, err := length.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal string length: %v", err)
	}

	// Combine length and string data
	result := append(lenBytes, strBytes...)
	return result, nil
}

func (s *String) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("string data is empty")
	}

	// Read the length
	var length VarInt
	if err := length.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal string length: %v", err)
	}

	// Calculate where the VarInt ends
	varIntSize := 0
	for _, b := range data {
		varIntSize++
		if b&0x80 == 0 {
			break
		}
	}

	// Check if we have enough data
	if len(data) < varIntSize+int(length) {
		return fmt.Errorf("string data too short")
	}

	// Extract the string
	s.Value = string(data[varIntSize : varIntSize+int(length)])
	return nil
}
