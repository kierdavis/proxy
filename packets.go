package proxy

type HS0HandshakePacket struct {
	ProtocolVersion uint64
	ServerAddress string
	ServerPort uint16
	NextState uint64
}

func (packet *HS0HandshakePacket) ID() (id PacketID) {
	return PacketID{Handshaking, Serverbound, 0x0}
}

func (packet *HS0HandshakePacket) Read(r BinaryReader) {
	packet.ProtocolVersion, _ = r.ReadVarint()
	packet.ServerAddress, _   = r.ReadString()
	packet.ServerPort, _      = r.ReadUint16()
	packet.NextState, _       = r.ReadVarint()
}

func (packet *HS0HandshakePacket) Write(w BinaryWriter) {
	w.WriteVarint(packet.ProtocolVersion)
	w.WriteString(packet.ServerAddress)
	w.WriteUint16(packet.ServerPort)
	w.WriteVarint(packet.NextState)
}

type PC40DisconnectPacket struct {
	JsonData string
}

func (packet *PC40DisconnectPacket) ID() (id PacketID) {
	return PacketID{Play, Clientbound, 0x40}
}

func (packet *PC40DisconnectPacket) Read(r BinaryReader) {
	packet.JsonData, _ = r.ReadString()
}

func (packet *PC40DisconnectPacket) Write(w BinaryWriter) {
	w.WriteString(packet.JsonData)
}

type LC0DisconnectPacket struct {
	JsonData string
}

func (packet *LC0DisconnectPacket) ID() (id PacketID) {
	return PacketID{Login, Clientbound, 0x40}
}

func (packet *LC0DisconnectPacket) Read(r BinaryReader) {
	packet.JsonData, _ = r.ReadString()
}

func (packet *LC0DisconnectPacket) Write(w BinaryWriter) {
	w.WriteString(packet.JsonData)
}

type LC1EncryptionRequestPacket struct {
	ServerID string
	PublicKey []byte
	VerifyToken []byte
}

func (packet *LC1EncryptionRequestPacket) ID() (id PacketID) {
	return PacketID{Login, Clientbound, 0x1}
}

func (packet *LC1EncryptionRequestPacket) Read(r BinaryReader) {
	packet.ServerID, _    = r.ReadString()
	publicKeyLength, _   := r.ReadUint16()
	packet.PublicKey, _   = r.ReadBytes(int(publicKeyLength))
	verifyTokenLength, _ := r.ReadUint16()
	packet.VerifyToken, _ = r.ReadBytes(int(verifyTokenLength))
}

func (packet *LC1EncryptionRequestPacket) Write(w BinaryWriter) {
	w.WriteString(packet.ServerID)
	w.WriteUint16(uint16(len(packet.PublicKey)))
	w.WriteBytes(packet.PublicKey)
	w.WriteUint16(uint16(len(packet.VerifyToken)))
	w.WriteBytes(packet.VerifyToken)
}

type LC2LoginSuccessPacket struct {
	UUID string
	Username string
}

func (packet *LC2LoginSuccessPacket) ID() (id PacketID) {
	return PacketID{Login, Clientbound, 0x2}
}

func (packet *LC2LoginSuccessPacket) Read(r BinaryReader) {
	packet.UUID, _     = r.ReadString()
	packet.Username, _ = r.ReadString()
}

func (packet *LC2LoginSuccessPacket) Write(w BinaryWriter) {
	w.WriteString(packet.UUID)
	w.WriteString(packet.Username)
}

type LS0LoginStartPacket struct {
	Name string
}

func (packet *LS0LoginStartPacket) ID() (id PacketID) {
	return PacketID{Login, Serverbound, 0x0}
}

func (packet *LS0LoginStartPacket) Read(r BinaryReader) {
	packet.Name, _ = r.ReadString()
}

func (packet *LS0LoginStartPacket) Write(w BinaryWriter) {
	w.WriteString(packet.Name)
}

type LS1EncryptionResponsePacket struct {
	EncryptedSharedSecret []byte
	EncryptedVerifyToken []byte
}

func (packet *LS1EncryptionResponsePacket) ID() (id PacketID) {
	return PacketID{Login, Serverbound, 0x1}
}

func (packet *LS1EncryptionResponsePacket) Read(r BinaryReader) {
	sharedSecretLength, _          := r.ReadUint16()
	packet.EncryptedSharedSecret, _ = r.ReadBytes(int(sharedSecretLength))
	verifyTokenLength, _           := r.ReadUint16()
	packet.EncryptedVerifyToken, _  = r.ReadBytes(int(verifyTokenLength))
}

func (packet *LS1EncryptionResponsePacket) Write(w BinaryWriter) {
	w.WriteUint16(uint16(len(packet.EncryptedSharedSecret)))
	w.WriteBytes(packet.EncryptedSharedSecret)
	w.WriteUint16(uint16(len(packet.EncryptedVerifyToken)))
	w.WriteBytes(packet.EncryptedVerifyToken)
}
