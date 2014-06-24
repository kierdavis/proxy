package proxy

import (
	"log"
	"net"
)

type Proxy struct {
	Errors chan error
	listener net.Listener
	bindAddr Address
	serverAddr Address
	hm *handlerManager
	gem *globalEncryptionManager
}

func New(bindAddr, serverAddr Address, username, password string) (proxy *Proxy, err error) {
	gem, err := newGlobalEncryptionManager(username, password)
	if err != nil {
		return nil, err
	}
	
	proxy = &Proxy{
		Errors: make(chan error, 10),
		bindAddr: bindAddr,
		serverAddr: serverAddr,
		hm: newHandlerManager(),
		gem: gem,
	}
	
	return proxy, nil
}

func (proxy *Proxy) AddHandler(handler interface{}) {
	proxy.hm.Add(handler)
}

func (proxy *Proxy) Run() (err error) {
	go proxy.RunAsync()
	
	return <-proxy.Errors
}

func (proxy *Proxy) RunAsync() {
	defer close(proxy.Errors)
	
	log.Printf("Listening on %s", proxy.bindAddr.String())
	
	ln, err := net.Listen("tcp", proxy.bindAddr.String())
	if err != nil {
		proxy.Errors <- err
		return
	}
	
	for {
		conn, err := ln.Accept()
		if err != nil {
			proxy.Errors <- err
			return
		}
		
		go proxy.handleConnection(conn)
	}
}

func (proxy *Proxy) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	
	log.Printf("Recieved connection from %s", clientConn.RemoteAddr().String())
	log.Printf("Connecting to %s", proxy.serverAddr.String())
	
	serverConn, err := net.Dial("tcp", proxy.serverAddr.String())
	if err != nil {
		log.Printf("Session error: %s", err.Error())
		return
	}
	defer serverConn.Close()
	
	sess := newSession(proxy, clientConn, serverConn)
	if sess == nil {
		return
	}
	
	err = sess.Run()
	if err != nil {
		log.Printf("Session error: %s", err.Error())
	}
}
