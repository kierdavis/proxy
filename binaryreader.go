package proxy

import (
	"encoding/binary"
	"io"
)

type BinaryReader struct {
	r io.Reader
}

func NewBinaryReader(r io.Reader) (br BinaryReader) {
	return BinaryReader{r}
}

func (br BinaryReader) ReadByte() (x byte, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadUint8() (x uint8, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadUint16() (x uint16, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadUint32() (x uint32, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadUint64() (x uint64, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadInt8() (x int8, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadInt16() (x int16, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadInt32() (x int32, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadInt64() (x int64, err error) {
	err = binary.Read(br.r, binary.BigEndian, &x)
	return x, err
}

func (br BinaryReader) ReadVarint() (x uint64, err error) {
	return binary.ReadUvarint(br)
}

func (br BinaryReader) ReadBytes(count int) (buf []byte, err error) {
	buf = make([]byte, count)
	pos := 0
	
	for pos < count {
		n, err := br.r.Read(buf[pos:])
		if err != nil {
			return nil, err
		}
		pos += n
	}
	
	return buf, nil
}

func (br BinaryReader) ReadString() (s string, err error) {
	length, err := br.ReadVarint()
	if err != nil {
		return "", err
	}
	
	buf, err := br.ReadBytes(int(length))
	if err != nil {
		return "", err
	}
	
	return string(buf), nil
}

func (br BinaryReader) ReadPacket() (p []byte, err error) {
	length, err := br.ReadVarint()
	if err != nil {
		return nil, err
	}
	
	body, err := br.ReadBytes(int(length))
	if err != nil {
		return nil, err
	}
	
	return body, nil
}

func (br BinaryReader) ReadSlot() (slot *Slot, err error) {
	item, err := br.ReadUint16()
	if err != nil {
		return nil, err
	}
	
	if item == 0xffff {
		return nil, nil
	}
	
	count, err := br.ReadUint8()
	if err != nil {
		return nil, err
	}
	
	damage, err := br.ReadUint16()
	if err != nil {
		return nil, err
	}
	
	nbtLength, err := br.ReadUint16()
	if err != nil {
		return nil, err
	}
	
	var nbt []byte
	
	if nbtLength != 0xffff {
		nbt, err = br.ReadBytes(int(nbtLength))
		if err != nil {
			return nil, err
		}
	}
	
	return &Slot{
		Item: item,
		Count: count,
		Damage: damage,
		NBT: nbt,
	}, nil
}
