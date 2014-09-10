package proxy

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

import crand "crypto/rand"

type serverEncryptionManager struct {
	gem *globalEncryptionManager
	accessToken string
	clientToken string
	selectedProfileID string
	remoteServerID string
	remotePublicKeyBytes []byte
	remotePublicKey rsa.PublicKey
	verifyToken []byte
	sharedSecret []byte
}

func (gem *globalEncryptionManager) newServer() (sem *serverEncryptionManager, err error) {
	sem = &serverEncryptionManager{
		gem: gem,
	}
	
	return sem, nil
}

func (sem *serverEncryptionManager) authenticate() (err error) {
	log.Printf("authenticate")
	
	requestMessage := authenticateRequest{
		Agent: authenticateAgent{
			Name: "Minecraft",
			Version: 1,
		},
		Username: sem.gem.username,
		Password: sem.gem.password,
	}
	
	requestJson, err := json.Marshal(requestMessage)
	if err != nil {
		return err
	}
	
	resp, err := http.Post("https://authserver.mojang.com/authenticate", "application/json", bytes.NewReader(requestJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP authentication error: %s", resp.Status)
	}
	
	var responseMessage authenticateResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&responseMessage)
	if err != nil {
		return err
	}
	
	sem.accessToken = responseMessage.AccessToken
	sem.clientToken = responseMessage.ClientToken
	sem.selectedProfileID = responseMessage.SelectedProfile.ID
	
	return nil
}

func (sem *serverEncryptionManager) handleEncryptionRequest(packet *LC1EncryptionRequestPacket) (err error) {
	sem.remoteServerID = packet.ServerID
	sem.remotePublicKeyBytes = packet.PublicKey
	sem.verifyToken = packet.VerifyToken
	
	sem.remotePublicKey, err = decodePublicKey(sem.remotePublicKeyBytes)
	if err != nil {
		return err
	}
	
	return nil
}

func (sem *serverEncryptionManager) generateSharedSecret() (err error) {
	//sem.sharedSecret = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 18}
	//return nil
	sem.sharedSecret = make([]byte, 16)
	_, err = io.ReadFull(crand.Reader, sem.sharedSecret)
	return err
}

func (sem *serverEncryptionManager) notifyJoin() (err error) {
	log.Printf("notifyJoin")
	
	serverHash := AuthDigest(sem.remoteServerID, sem.sharedSecret, sem.remotePublicKeyBytes)
	requestMessage := notifyJoinRequest{
		AccessToken: sem.accessToken,
		SelectedProfile: sem.selectedProfileID,
		ServerID: serverHash,
	}
	
	requestJson, err := json.Marshal(requestMessage)
	if err != nil {
		return err
	}
	
	resp, err := http.Post("https://sessionserver.mojang.com/session/minecraft/join", "application/json", bytes.NewReader(requestJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP authentication error: %s", resp.Status)
	}
	
	return nil
}

func (sem *serverEncryptionManager) makeEncryptionResponse() (packet *LS1EncryptionResponsePacket, err error) {
	encryptedSharedSecret, err := rsa.EncryptPKCS1v15(crand.Reader, &sem.remotePublicKey, sem.sharedSecret)
	if err != nil {
		return nil, err
	}
	
	encryptedVerifyToken, err := rsa.EncryptPKCS1v15(crand.Reader, &sem.remotePublicKey, sem.verifyToken)
	if err != nil {
		return nil, err
	}
	
	packet = &LS1EncryptionResponsePacket{
		EncryptedSharedSecret: encryptedSharedSecret,
		EncryptedVerifyToken: encryptedVerifyToken,
	}
	
	return packet, nil
}

type authenticateRequest struct {
	Agent authenticateAgent `json:"agent"`
	Username string `json:"username"`
	Password string `json:"password"`
	ClientToken string `json:"clientToken,omitempty"`
}

type authenticateAgent struct {
	Name string `json:"name"`
	Version int `json:"version"`
}

type authenticateResponse struct {
	AccessToken string `json:"accessToken"`
	ClientToken string `json:"clientToken"`
	AvailableProfiles []authenticateProfile `json:"availableProfiles"`
	SelectedProfile authenticateProfile `json:"selectedProfile"`
}

type authenticateProfile struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Legacy bool `json:"legacy,omitempty"`
}

type notifyJoinRequest struct {
	AccessToken string `json:"accessToken"`
	SelectedProfile string `json:"selectedProfile"`
	ServerID string `json:"serverId"`
}
