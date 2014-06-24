package proxy

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

import crand "crypto/rand"

type clientEncryptionManager struct {
	gem *globalEncryptionManager
	playerName string
	playerUUID string
	verifyToken []byte
	sharedSecret []byte
}

func (gem *globalEncryptionManager) newClient(playerName string) (cem *clientEncryptionManager, err error) {
	verifyToken := make([]byte, 16)
	_, err = io.ReadFull(crand.Reader, verifyToken)
	if err != nil {
		return nil, err
	}
	
	cem = &clientEncryptionManager{
		gem: gem,
		playerName: playerName,
		verifyToken: verifyToken,
	}
	
	return cem, nil
}

func (cem *clientEncryptionManager) makeEncryptionRequest() (packet *LC1EncryptionRequestPacket, err error) {
	packet = &LC1EncryptionRequestPacket{
		ServerID: cem.gem.serverID,
		PublicKey: cem.gem.encodedPublicKey,
		VerifyToken: cem.verifyToken,
	}
	
	return packet, nil
}

func (cem *clientEncryptionManager) handleEncryptionResponse(packet *LS1EncryptionResponsePacket) (err error) {
	cem.sharedSecret, err = rsa.DecryptPKCS1v15(crand.Reader, cem.gem.privateKey, packet.EncryptedSharedSecret)
	if err != nil {
		return err
	}
	
	returnedVerifyToken, err := rsa.DecryptPKCS1v15(crand.Reader, cem.gem.privateKey, packet.EncryptedVerifyToken)
	if err != nil {
		return err
	}
	
	if len(cem.verifyToken) != len(returnedVerifyToken) {
		return fmt.Errorf("Authentication failure")
	}
	
	var diff uint8
	for i := 0; i < len(cem.verifyToken); i++ {
		diff |= cem.verifyToken[i] ^ returnedVerifyToken[i]
	}
	
	if diff != 0 {
		return fmt.Errorf("Authentication failure")
	}
	
	return nil
}

func (cem *clientEncryptionManager) notifyHasJoined() (err error) {
	log.Printf("notifyHasJoined")
	
	serverHash := AuthDigest(cem.gem.serverID, cem.sharedSecret, cem.gem.encodedPublicKey)
	
	params := make(url.Values)
	params.Set("username", cem.playerName)
	params.Set("serverId", serverHash)
	
	resp, err := http.Get("https://sessionserver.mojang.com/session/minecraft/hasJoined?" + params.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP authentication error: %s", resp.Status)
	}
	
	var responseMessage hasJoinedResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&responseMessage)
	if err != nil {
		return err
	}
	
	cem.playerUUID = responseMessage.UUID
	
	return nil
}

type hasJoinedResponse struct {
	UUID string `json:"id"`
	Properties []hasJoinedProperty `json:"properties"`
}

type hasJoinedProperty struct {
	Name string `json:"name"`
	Value interface{} `json:"value"`
}
