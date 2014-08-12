// This program demonstrates using the library to run a proxy that intercepts
// chat message packets sent to the server.
package main

import (
    "github.com/kierdavis/proxy"
    "log"
    "os"
    "strings"
)

// TCP address on which the proxy server will listen.
var ListenAddr = proxy.Address{"localhost", 25567}

// TCP address that connections will be forwarded to.
var ServerAddr = proxy.Address{"localhost", 25565}

// Minecraft account login details.
// Don't forget to use your email address for the username if you are using a
// Mojang account.
const Username = "username"
const Password = "password"

func main() {
	// Initialise the proxy.
	prox, err := proxy.New(ListenAddr, ServerAddr, Username, Password)
	if err != nil {
		log.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
	
	// Add any packet handlers to the proxy.
	prox.AddHandler(chatPacketHandler)
	
	// Start the proxy server.
	err = prox.Run()
	if err != nil {
		log.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

// An example packet handler.
// All packet handlers are a function with two arguments:
//   - a value of type *proxy.Session that encapsulates the connections to the
//     client and the server.
//   - a value of any type implementing proxy.Packet that contains information
//     about the packet being processed. It may be modified by the packet
//     handler.
// The type of the second argument determines which packets the handler will be
// triggered for.
// Packet handlers must return a value of type bool. If this value is true, the
// modified packet continues on its journey to the server. If it is false, the
// packet is dropped.
func chatPacketHandler(session *proxy.Session, packet *PS1ChatMessagePacket) bool {
	// This code will be run whenever the Minecraft client sends a chat message
	// packet to the server i.e. whenever the player types a chat message and
	// presses enter.
	
	// We will check to see if the player has sent a message beginning with
	// "/greet".
	if strings.HasPrefix(packet.Message, "/greet") {
		// We will then construct a new packet, this one a clientbound chat
		// message packet. It has a slightly different format to the serverbound
		// chat packet we just received.
		newPacket := &PC2ChatMessagePacket{}
		
		// The message field in the clientbound chat packet is a JSON string,
		// unlike the serverbound chat packet.
		newPacket.JsonData = "{'text':'Hello from the proxy!','color':'red'}"
		
		// Send this packet to the client.
		session.Send(newPacket)
		
		// Return false so that the original packet never reaches the server (if
		// it did, the server would not recognise the /greet command and would
		// produce an error).
		return false
	}
	
	// Otherwise, return true to allow the packet to continue to the server.
	return true
}

// We need to define the packet types used in the above handler.
// In the Minecraft protocol (as of Minecraft version 1.7), each type of packet
// is identified by three components:
//   * The "protocol state" - one of Handshaking, Play, Status or Login.
//   * The direction - whether the packet is serverbound (sent from the client
//     to the server) or clientbound (sent from the server to the client).
//   * The number - an integer uniquely identifying the packet type within each
//     combination of state and direction.
// http://wiki.vg/Protocol has a comprehensive list of all packet types in the
// latest version of Minecraft.

// A naming convention used in this package is for packet type names to have the
// format "<state><direction><number (hexadecimal)><name>Packet", where <state>, <direction>
// and <number> are the elements of the packet ID and are unique to each type of
// packet, and <name> is the name of the packet type as specified on the list at
// http://wiki.vg/Protocol.
// The reason for prefixing packet type names with characters identifying the ID
// is to distinguish packets with the same name but different states or
// directions (for example, the clientbound and serverbound chat messages are
// different).

// First up, the serverbound Chat Message packet. It contains only one field,
// which is a string. Its name is prefixed with "PS1", since it is a Play-state
// packet, it is Serverbound, and within the domain of Play-state serverbound
// packets it is numbered 1.
type PS1ChatMessagePacket struct {
	Message string
}

// All packet types must define an ID method that returns the packet's ID as a
// proxy.PacketID value.
func (packet *PS1ChatMessagePacket) ID() proxy.PacketID {
	return proxy.PacketID{proxy.Play, proxy.Serverbound, 0x1}
}

// They must also define a Read method, to read the packet's values from the
// binary data stream, and a Write method to do the opposite.
func (packet *PS1ChatMessagePacket) Read(r proxy.BinaryReader) {
	// The second return value from r.Read*() will always be nil in this
	// context, so we can ignore it.
	packet.Message, _ = r.ReadString()
}

func (packet *PS1ChatMessagePacket) Write(w proxy.BinaryWriter) {
	w.WriteString(packet.Message)
}

// We will also define the clientbound Chat Message packet. It is a Play-state,
// Clientbound packet with number 2, and contains a single string-type field
// containing the JSON data describing the chat message.
type PC2ChatMessagePacket struct {
	JsonData string
}

func (packet *PC2ChatMessagePacket) ID() proxy.PacketID {
	return proxy.PacketID{proxy.Play, proxy.Clientbound, 0x2}
}

func (packet *PC2ChatMessagePacket) Read(r proxy.BinaryReader) {
	packet.JsonData, _ = r.ReadString()
}

func (packet *PC2ChatMessagePacket) Write(w proxy.BinaryWriter) {
	w.WriteString(packet.JsonData)
}
