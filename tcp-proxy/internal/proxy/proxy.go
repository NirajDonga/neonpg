package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Proxy struct {
	listenAddr string
	namespace  string
}

func NewProxy(listenAddr, namespace string) *Proxy {
	return &Proxy{listenAddr: listenAddr, namespace: namespace}
}

func (p *Proxy) Start() error {
	ln, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return fmt.Errorf("proxy listen: %w", err)
	}
	log.Printf("[Proxy] Listening on %s", p.listenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[Proxy] Accept error: %v", err)
			continue
		}
		go p.handle(conn)
	}
}

func (p *Proxy) handle(client net.Conn) {
	defer client.Close()

	startupMsg, err := readStartupMessage(client)
	if err != nil {
		log.Printf("[Proxy] Failed to read startup: %v", err)
		return
	}

	tenantID, err := extractUser(startupMsg)
	if err != nil {
		log.Printf("[Proxy] Failed to extract user: %v", err)
		return
	}

	upstream := fmt.Sprintf("%s-svc.%s.svc.cluster.local:5432", tenantID, p.namespace)
	log.Printf("[Proxy] Routing %s → %s", tenantID, upstream)

	server, err := net.Dial("tcp", upstream)
	if err != nil {
		log.Printf("[Proxy] Failed to connect to %s: %v", upstream, err)
		return
	}
	defer server.Close()

	if _, err := server.Write(startupMsg); err != nil {
		log.Printf("[Proxy] Failed to forward startup: %v", err)
		return
	}

	// bidirectional copy
	done := make(chan struct{}, 2)
	go func() { io.Copy(server, client); done <- struct{}{} }()
	go func() { io.Copy(client, server); done <- struct{}{} }()
	<-done
}

func readStartupMessage(conn net.Conn) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, fmt.Errorf("read length: %w", err)
	}

	msgLen := int(binary.BigEndian.Uint32(lenBuf))
	if msgLen < 8 || msgLen > 10000 {
		return nil, fmt.Errorf("invalid message length: %d", msgLen)
	}

	rest := make([]byte, msgLen-4)
	if _, err := io.ReadFull(conn, rest); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return append(lenBuf, rest...), nil
}

func extractUser(msg []byte) (string, error) {
	if len(msg) < 9 {
		return "", fmt.Errorf("message too short")
	}

	parts := strings.Split(string(msg[8:]), "\x00")
	for i := 0; i+1 < len(parts); i += 2 {
		if parts[i] == "user" {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("no user field in startup message")
}
