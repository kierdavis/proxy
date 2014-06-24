package proxy

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

type Session struct {
	Proxy *Proxy
	
	clientConn net.Conn
	serverConn net.Conn
	
	clientCodec *codec
	serverCodec *codec
	
	state State
	cem *clientEncryptionManager
	sem *serverEncryptionManager
	
	// Handshake info
	ProtocolVersion uint64
	handshakeNextState uint64
	
	// Login info
	PlayerName string
	UUID string
	
	outgoingChan chan Packet
}

func newSession(proxy *Proxy, clientConn net.Conn, serverConn net.Conn) (s *Session) {
	s = &Session{
		Proxy: proxy,
		clientConn: clientConn,
		serverConn: serverConn,
		clientCodec: newCodec(clientConn),
		serverCodec: newCodec(serverConn),
		state: Handshaking,
	}
	
	return s
}

func (s *Session) Run() (err error) {
	err = s.run()
	if err != nil {
		switch s.state {
		case Play:
			s.send(&PC40DisconnectPacket{"{\"text\":\"Internal proxy error\"}"})
		
		case Login:
			s.send(&LC0DisconnectPacket{"{\"text\":\"Internal proxy error\"}"})
		}
	}
	
	return err
}

func (s *Session) run() (err error) {
	err = s.passHandshake()
	if err != nil {
		return err
	}
	
	switch s.handshakeNextState {
	case 1:
		return s.doStatus()
	case 2:
		return s.doLogin()
	}
	
	return fmt.Errorf("Invalid handshake next state %d", s.handshakeNextState)
}

func (s *Session) doStatus() (err error) {
	s.setState(Status)
	return s.passPackets()
}

func (s *Session) doLogin() (err error) {
	s.setState(Login)
	
	err = s.passLoginStart()
	if err != nil {
		return err
	}
	
	/*
	cem, err := s.Proxy.gem.newClient(s.PlayerName)
	if err != nil {
		return err
	}
	*/
	
	sem, err := s.Proxy.gem.newServer()
	if err != nil {
		return err
	}
	
	//s.cem = cem
	s.sem = sem
	
	err = s.sem.authenticate()
	if err != nil {
		return err
	}
	
	err = s.readEncryptionRequest()
	if err != nil {
		return err
	}
	
	/*
	err = s.writeEncryptionRequest()
	if err != nil {
		return err
	}
	*/
	
	err = s.sem.generateSharedSecret()
	if err != nil {
		return err
	}
	
	err = s.sem.notifyJoin()
	if err != nil {
		return err
	}
	
	/*
	err = s.readEncryptionResponse()
	if err != nil {
		return err
	}
	*/
	
	err = s.writeEncryptionResponse()
	if err != nil {
		return err
	}
	
	/*
	err = s.cem.notifyHasJoined()
	if err != nil {
		return err
	}
	*/
	
	/*
	err = s.clientCodec.Encrypt(s.cem.sharedSecret)
	if err != nil {
		return err
	}
	*/
	
	err = s.serverCodec.Encrypt(s.sem.sharedSecret)
	if err != nil {
		return err
	}
	
	err = s.passLoginSuccess()
	if err != nil {
		return err
	}
	
	log.Printf("Login successful")
	
	s.setState(Play)
	return s.passPackets()
}

func (s *Session) passPackets() (err error) {
	errs := make(chan error, 1)
	clientIncoming := make(chan []byte, 10)
	serverIncoming := make(chan []byte, 10)
	clientOutgoing := make(chan []byte, 10)
	serverOutgoing := make(chan []byte, 10)
	outgoing := make(chan Packet, 10)
	
	s.outgoingChan = outgoing
	
	go s.clientCodec.ReadAll(clientIncoming, errs)
	go s.serverCodec.ReadAll(serverIncoming, errs)
	go s.clientCodec.WriteAll(clientOutgoing, errs)
	go s.serverCodec.WriteAll(serverOutgoing, errs)
	
	go func() {
		for {
			select {
			case packetData := <-clientIncoming:
				packetData, dir, accept := s.handlePacket(packetData, Serverbound)
				if accept {
					switch dir {
					case Clientbound:
						clientOutgoing <- packetData
					case Serverbound:
						serverOutgoing <- packetData
					}
				}
			
			case packetData := <-serverIncoming:
				packetData, dir, accept := s.handlePacket(packetData, Clientbound)
				if accept {
					switch dir {
					case Clientbound:
						clientOutgoing <- packetData
					case Serverbound:
						serverOutgoing <- packetData
					}
				}
			
			case packet := <-outgoing:
				buf := bytes.NewBuffer(nil)
				w := NewBinaryWriter(buf)
				w.WriteVarint(packet.ID().Number)
				packet.Write(w)
				
				switch packet.ID().Direction {
				case Clientbound:
					clientOutgoing <- buf.Bytes()
				case Serverbound:
					serverOutgoing <- buf.Bytes()
				}
			}
		}
	}()
	
	err = <-errs
	time.Sleep(time.Second / 2)
	return err
}

func (s *Session) handlePacket(packetData []byte, dir Direction) (newPacketData []byte, newDir Direction, accept bool) {
	r := NewBinaryReader(bytes.NewReader(packetData))
	idNum, _ := r.ReadVarint()
	id := PacketID{s.state, dir, idNum}
	
	packet := s.Proxy.hm.Lookup(id)
	if packet != nil {
		packet.Read(r)
		accept = s.Proxy.hm.Process(s, packet)
		if !accept {
			return packetData, dir, false
		}
		
		buf := bytes.NewBuffer(nil)
		w := NewBinaryWriter(buf)
		w.WriteVarint(packet.ID().Number)
		packet.Write(w)
		return buf.Bytes(), packet.ID().Direction, true
	}
	
	return packetData, dir, true
}

func (s *Session) passHandshake() (err error) {
	packet := &HS0HandshakePacket{}
	err = s.recv(packet)
	if err != nil {
		return err
	}
	
	s.ProtocolVersion = packet.ProtocolVersion
	s.handshakeNextState = packet.NextState
	
	packet.ServerAddress = s.Proxy.serverAddr.Host
	packet.ServerPort = uint16(s.Proxy.serverAddr.Port)
	
	return s.send(packet)
}

func (s *Session) passLoginStart() (err error) {
	packet := &LS0LoginStartPacket{}
	err = s.recv(packet)
	if err != nil {
		return err
	}
	
	s.PlayerName = packet.Name
	
	return s.send(packet)
}

func (s *Session) readEncryptionRequest() (err error) {
	packet := &LC1EncryptionRequestPacket{}
	err = s.recv(packet)
	if err != nil {
		return err
	}
	
	return s.sem.handleEncryptionRequest(packet)
}

func (s *Session) writeEncryptionRequest() (err error) {
	packet, err := s.cem.makeEncryptionRequest()
	if err != nil {
		return err
	}
	
	return s.send(packet)
}

func (s *Session) readEncryptionResponse() (err error) {
	packet := &LS1EncryptionResponsePacket{}
	err = s.recv(packet)
	if err != nil {
		return err
	}
	
	return s.cem.handleEncryptionResponse(packet)
}

func (s *Session) writeEncryptionResponse() (err error) {
	packet, err := s.sem.makeEncryptionResponse()
	if err != nil {
		return err
	}
	
	return s.send(packet)
}

func (s *Session) passLoginSuccess() (err error) {
	packet := &LC2LoginSuccessPacket{}
	err = s.recv(packet)
	if err != nil {
		return err
	}
	
	s.UUID = packet.UUID
	s.PlayerName = packet.Username
	
	return s.send(packet)
}

func (s *Session) recv(packet Packet) (err error) {
	id := packet.ID()
	if s.state != id.State {
		panic(fmt.Sprintf("Session.recv: wrong state! (recving a %s packet in state %s)", id.State.String(), s.state.String()))
	}
	
	var c *codec
	switch id.Direction {
	case Clientbound:
		c = s.serverCodec
	case Serverbound:
		c = s.clientCodec
	}
	
	packetData, err := c.Read()
	if err != nil {
		return err
	}
	
	r := NewBinaryReader(bytes.NewReader(packetData))
	idNum, _ := r.ReadVarint()
	if idNum != id.Number {
		return fmt.Errorf("Unexpected %s:%s:%X packet (expecting %s:%s:%X)", id.State.String(), id.Direction.String(), idNum, id.State.String(), id.Direction.String(), id.Number)
	}
	
	packet.Read(r)
	
	//fmt.Printf("recv %#v\n", packet)
	
	return nil
}

func (s *Session) send(packet Packet) (err error) {
	//fmt.Printf("send %#v\n", packet)
	
	id := packet.ID()
	if s.state != id.State {
		panic(fmt.Sprintf("Session.send: wrong state! (sending a %s packet in state %s)", id.State.String(), s.state.String()))
	}
	
	var c *codec
	switch id.Direction {
	case Clientbound:
		c = s.clientCodec
	case Serverbound:
		c = s.serverCodec
	}
	
	buf := bytes.NewBuffer(nil)
	w := NewBinaryWriter(buf)
	w.WriteVarint(id.Number)
	packet.Write(w)
	
	err = c.Write(buf.Bytes())
	if err != nil {
		return err
	}
	
	return nil
}

func (s *Session) Send(packet Packet) {
	if s.outgoingChan != nil {
		s.outgoingChan <- packet
	} else {
		err := s.send(packet)
		if err != nil {
			log.Printf("Session error: %s", err.Error())
		}
	}
}

func (s *Session) setState(state State) {
	s.state = state
	log.Printf("Changing state: %s", state.String())
}
