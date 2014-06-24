package proxy

import (
	"encoding/binary"
	"io"
)

type BinaryWriter struct {
	w io.Writer
}

func NewBinaryWriter(w io.Writer) (br BinaryWriter) {
	return BinaryWriter{w}
}

func (br BinaryWriter) WriteByte(x byte) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteUint8(x uint8) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteUint16(x uint16) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteUint32(x uint32) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteUint64(x uint64) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteInt8(x int8) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteInt16(x int16) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteInt32(x int32) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteInt64(x int64) (err error) {
	return binary.Write(br.w, binary.BigEndian, x)
}

func (br BinaryWriter) WriteVarint(x uint64) (err error) {
	buf := make([]byte, 32)
	n := binary.PutUvarint(buf, x)
	return br.WriteBytes(buf[:n])
}

func (br BinaryWriter) WriteBytes(buf []byte) (err error) {
	for len(buf) > 0 {
		n, err := br.w.Write(buf)
		if err != nil {
			return err
		}
		buf = buf[n:]
	}
	
	return nil
}

func (br BinaryWriter) WriteString(s string) (err error) {
	err = br.WriteVarint(uint64(len(s)))
	if err != nil {
		return err
	}
	
	return br.WriteBytes([]byte(s))
}

func (br BinaryWriter) WritePacket(p []byte) (err error) {
	err = br.WriteVarint(uint64(len(p)))
	if err != nil {
		return err
	}
	
	return br.WriteBytes(p)
}
