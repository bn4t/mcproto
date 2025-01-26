package types

type Boolean bool

func (b Boolean) Marshal() ([]byte, error) {
	if b {
		return []byte{1}, nil
	}
	return []byte{0}, nil
}

func (b *Boolean) Unmarshal(data []byte) error {
	*b = data[0] == 1
	return nil
}
