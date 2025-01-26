package types

type VarInt int32

func (v VarInt) Marshal() ([]byte, error) {
	// TODO
}

func (v *VarInt) Unmarshal(data []byte) error {
	// TODO
}
