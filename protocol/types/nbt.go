package types

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
)

type NBTTag byte

const (
	TagEnd       NBTTag = 0
	TagByte      NBTTag = 1
	TagShort     NBTTag = 2
	TagInt       NBTTag = 3
	TagLong      NBTTag = 4
	TagFloat     NBTTag = 5
	TagDouble    NBTTag = 6
	TagByteArray NBTTag = 7
	TagString    NBTTag = 8
	TagList      NBTTag = 9
	TagCompound  NBTTag = 10
	TagIntArray  NBTTag = 11
	TagLongArray NBTTag = 12
)

type NBTValue interface{}

type NBTCompound struct {
	Name  string
	Value map[string]NBTValue
}

type NBT struct {
	Root *NBTCompound
}

func (n NBT) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if err := n.writeTag(gzipWriter, TagCompound, n.Root); err != nil {
		gzipWriter.Close()
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (n *NBT) Unmarshal(data []byte) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	reader := bufio.NewReader(gzipReader)

	tagType, err := reader.ReadByte()
	if err != nil {
		return err
	}

	if NBTTag(tagType) != TagCompound {
		return fmt.Errorf("expected root compound tag, got %d", tagType)
	}

	compound, err := n.readCompound(reader)
	if err != nil {
		return err
	}

	n.Root = compound
	return nil
}

func (n NBT) writeTag(w io.Writer, tagType NBTTag, value interface{}) error {
	switch tagType {
	case TagEnd:
		return binary.Write(w, binary.BigEndian, byte(0))
	case TagByte:
		return binary.Write(w, binary.BigEndian, value.(byte))
	case TagShort:
		return binary.Write(w, binary.BigEndian, value.(int16))
	case TagInt:
		return binary.Write(w, binary.BigEndian, value.(int32))
	case TagLong:
		return binary.Write(w, binary.BigEndian, value.(int64))
	case TagFloat:
		return binary.Write(w, binary.BigEndian, value.(float32))
	case TagDouble:
		return binary.Write(w, binary.BigEndian, value.(float64))
	case TagString:
		str := value.(string)
		if err := binary.Write(w, binary.BigEndian, int16(len(str))); err != nil {
			return err
		}
		_, err := w.Write([]byte(str))
		return err
	case TagCompound:
		compound := value.(*NBTCompound)
		if err := n.writeString(w, compound.Name); err != nil {
			return err
		}
		for name, val := range compound.Value {
			if err := n.writeNamedTag(w, name, val); err != nil {
				return err
			}
		}
		return binary.Write(w, binary.BigEndian, byte(TagEnd))
	default:
		return fmt.Errorf("unsupported tag type: %d", tagType)
	}
}

func (n NBT) writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.BigEndian, int16(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func (n NBT) writeNamedTag(w io.Writer, name string, value interface{}) error {
	var tagType NBTTag
	switch value.(type) {
	case byte:
		tagType = TagByte
	case int16:
		tagType = TagShort
	case int32:
		tagType = TagInt
	case int64:
		tagType = TagLong
	case float32:
		tagType = TagFloat
	case float64:
		tagType = TagDouble
	case string:
		tagType = TagString
	case *NBTCompound:
		tagType = TagCompound
	default:
		return fmt.Errorf("unsupported value type for tag")
	}

	if err := binary.Write(w, binary.BigEndian, byte(tagType)); err != nil {
		return err
	}

	if err := n.writeString(w, name); err != nil {
		return err
	}

	return n.writeTag(w, tagType, value)
}

func (n *NBT) readCompound(r io.Reader) (*NBTCompound, error) {
	name, err := n.readString(r)
	if err != nil {
		return nil, err
	}

	compound := &NBTCompound{
		Name:  name,
		Value: make(map[string]NBTValue),
	}

	reader := bufio.NewReader(r)
	for {
		tagType, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		if NBTTag(tagType) == TagEnd {
			break
		}

		name, err := n.readString(r)
		if err != nil {
			return nil, err
		}

		value, err := n.readTag(r, NBTTag(tagType))
		if err != nil {
			return nil, err
		}

		compound.Value[name] = value
	}

	return compound, nil
}

func (n *NBT) readString(r io.Reader) (string, error) {
	var length int16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}

	bytes := make([]byte, length)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (n *NBT) readTag(r io.Reader, tagType NBTTag) (interface{}, error) {
	switch tagType {
	case TagByte:
		var v byte
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagShort:
		var v int16
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagInt:
		var v int32
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagLong:
		var v int64
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagFloat:
		var v float32
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagDouble:
		var v float64
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagString:
		return n.readString(r)
	case TagCompound:
		return n.readCompound(r)
	default:
		return nil, fmt.Errorf("unsupported tag type: %d", tagType)
	}
}
