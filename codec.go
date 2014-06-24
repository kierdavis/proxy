package proxy

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"github.com/kierdavis/cfb8"
	"io"
	"log"
)

type codec struct {
	conn io.ReadWriter
	bufr *bufio.Reader
	bufw *bufio.Writer
	binr BinaryReader
	binw BinaryWriter
}

func newCodec(conn io.ReadWriter) (c *codec) {
	c = &codec{conn: conn}
	c.bufr = bufio.NewReader(c.conn)
	c.bufw = bufio.NewWriter(c.conn)
	c.binr = NewBinaryReader(c.bufr)
	c.binw = NewBinaryWriter(c.bufw)
	
	return c
}

func (c *codec) Read() (packet []byte, err error) {
	return c.binr.ReadPacket()
}

func (c *codec) ReadAll(packetChan chan []byte, errChan chan error) {
	for {
		packet, err := c.Read()
		if err != nil {
			errChan <- err
			return
		}
		
		packetChan <- packet
	}
}

func (c *codec) Write(packet []byte) (err error) {
	err = c.binw.WritePacket(packet)
	if err != nil {
		return err
	}
	return c.bufw.Flush()
}

func (c *codec) WriteAll(packetChan chan []byte, errChan chan error) {
	for packet := range packetChan {
		err := c.Write(packet)
		if err != nil {
			errChan <- err
			return
		}
	}
}

func (c *codec) Encrypt(sharedSecret []byte) (err error) {
	log.Printf("Enabling encryption")
	
	decCipher, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return err
	}
	
	encCipher, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return err
	}
	
	dec := cfb8.NewDecrypter(decCipher, sharedSecret)
	enc := cfb8.NewEncrypter(encCipher, sharedSecret)
	
	decr := cipher.StreamReader{dec, c.bufr}
	encw := cipher.StreamWriter{enc, c.bufw, nil}
	
	c.binr = NewBinaryReader(decr)
	c.binw = NewBinaryWriter(encw)
	
	return nil
}
