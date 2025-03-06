package types

import (
	"encoding/json"
	"fmt"
)

type ChatComponent struct {
	Text          string          `json:"text,omitempty"`
	Color         string          `json:"color,omitempty"`
	Bold          *bool           `json:"bold,omitempty"`
	Italic        *bool           `json:"italic,omitempty"`
	Underlined    *bool           `json:"underlined,omitempty"`
	Strikethrough *bool           `json:"strikethrough,omitempty"`
	Obfuscated    *bool           `json:"obfuscated,omitempty"`
	Extra         []ChatComponent `json:"extra,omitempty"`
}

type Chat ChatComponent

func (c Chat) Marshal() ([]byte, error) {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat: %v", err)
	}

	// Convert to string first to properly handle UTF-8
	str := string(jsonBytes)
	strBytes := []byte(str)

	// Create result with VarInt length prefix
	length := len(strBytes)
	varIntLen := VarInt(length)
	lenBytes, err := varIntLen.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal length: %v", err)
	}

	result := append(lenBytes, strBytes...)
	return result, nil
}

func (c *Chat) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("chat data is empty")
	}

	// Read length
	var length VarInt
	if err := length.Unmarshal(data); err != nil {
		return fmt.Errorf("failed to unmarshal length: %v", err)
	}

	// Calculate where the VarInt ends
	varIntSize := 0
	for _, b := range data {
		varIntSize++
		if b&0x80 == 0 {
			break
		}
	}

	// Extract the JSON string
	if len(data) < varIntSize+int(length) {
		return fmt.Errorf("chat data too short")
	}
	jsonData := data[varIntSize : varIntSize+int(length)]

	// Unmarshal the JSON
	if err := json.Unmarshal(jsonData, c); err != nil {
		return fmt.Errorf("failed to unmarshal chat json: %v", err)
	}

	return nil
}
