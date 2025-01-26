package types

type MarshallableType interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}
