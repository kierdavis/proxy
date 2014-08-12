package proxy

import (
	"fmt"
)

type State int

const (
	Handshaking State = iota
	Play
	Status
	Login
)

func (s State) String() string {
	switch s {
	case Handshaking:
		return "Handshaking"
	case Play:
		return "Play"
	case Status:
		return "Status"
	case Login:
		return "Login"
	}
	
	return ""
}

type Direction int

const (
	Clientbound Direction = iota
	Serverbound
)

func (d Direction) String() string {
	switch d {
	case Clientbound:
		return "Clientbound"
	case Serverbound:
		return "Serverbound"
	}
	
	return ""
}

type PacketID struct {
	State State
	Direction Direction
	Number uint64
}

func (id PacketID) String() (s string) {
	return fmt.Sprintf("%s:%s:%X", id.State.String(), id.Direction.String(), id.Number)
}

type Packet interface {
	ID() PacketID
	Read(BinaryReader)
	Write(BinaryWriter)
}

type Address struct {
	Host string
	Port int
}

func (addr Address) String() (s string) {
	return fmt.Sprintf("%s:%d", addr.Host, addr.Port)
}

type Slot struct {
	Item uint16
	Count uint8
	Damage uint16
	NBT []byte
}
