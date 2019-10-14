package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
)

//Server definition
type Server struct {
	serverVersion     string // e.g. "8.0.12"
	protocolVersion   int    // minimal 10
	capability        uint32 // server capability flag
	collationID       uint8
	defaultAuthMethod string // default authentication method, 'mysql_native_password'
	pubKey            []byte
	tlsConfig         *tls.Config
	cacheShaPassword  *sync.Map // 'user@host' -> SHA256(SHA256(PASSWORD))
}

//NewServer creates a new server
func NewServer(serverVersion string, protocolVersion int, collationID uint8, defaultAuthMethod string) *Server {
	const capability uint32 = 0
	return &Server{
		serverVersion,
		protocolVersion,
		capability,
		collationID,
		defaultAuthMethod,
		nil,
		nil,
		new(sync.Map),
	}
}

//Serve on requests
func Serve(server *Server, port int) error {
	service := fmt.Sprintf(":%d", port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handle(server, conn)
	}
}

func handle(server *Server, conn net.Conn) {
	//empty
}
