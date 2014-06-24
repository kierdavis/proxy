package proxy

import (
	"crypto/rsa"
	"io"
	"log"
)

import crand "crypto/rand"

type globalEncryptionManager struct {
	username string
	password string
	privateKey *rsa.PrivateKey
	encodedPublicKey []byte
	serverID string
}

func newGlobalEncryptionManager(username, password string) (gem *globalEncryptionManager, err error) {
	log.Printf("Generating keypair")
	
	privateKey, err := rsa.GenerateKey(crand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	
	encodedPublicKey, err := encodePublicKey(privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	
	serverIDBytes := make([]byte, 20)
	_, err = io.ReadFull(crand.Reader, serverIDBytes)
	if err != nil {
		return nil, err
	}
	
	for i := 0; i < len(serverIDBytes); i++ {
		// Characters must be in 0x21 - 0x7E, but lets use 0x30 - 0x6F for simplicity.
		serverIDBytes[i] = (serverIDBytes[i] & 0x3F) + 0x30
	}
	
	gem = &globalEncryptionManager{
		username: username,
		password: password,
		privateKey: privateKey,
		encodedPublicKey: encodedPublicKey,
		serverID: string(serverIDBytes),
	}
	
	return gem, nil
}
