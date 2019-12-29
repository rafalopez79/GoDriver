package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	util "github.com/rafalopez79/godriver/internal/util"
	mysql "github.com/rafalopez79/godriver/mysql"
)

//Server definition
type Server struct {
	serverVersion     string // e.g. "8.0.12"
	protocolVersion   byte   // minimal 10
	capability        uint32 // server capability flag
	collationID       uint8
	defaultAuthMethod string // default authentication method, 'mysql_native_password'
	pubKey            []byte
	tlsConfig         *tls.Config
	cacheShaPassword  *sync.Map        // 'user@host' -> SHA256(SHA256(PASSWORD))
	connectionCount   uint32           //conn id tracker
	sessions          *sync.Map        //[uint64]Session
	bufferPool        *util.BufferPool //bufferpool
	listener          *net.TCPListener //listener
}

//NewServer creates a new server
func NewServer(serverVersion string, defaultAuthMethod string) (server *Server, err error) {
	const capability uint32 = mysql.ClientLongPassword | mysql.ClientLongFlag | mysql.ClientConnectWithDB |
		mysql.ClientProtocol41 | mysql.ClientTransactions | mysql.ClientSecureConnection | mysql.ClientPluginAuth |
		mysql.ClientPluginAuthLENENCClientData | mysql.ClientCompress | mysql.ClientSSL
	caPem, caKey, err := generateCA()
	if err != nil {
		return nil, err
	}
	certPem, keyPem, err := generateAndSignRSACerts(caPem, caKey)
	if err != nil {
		return nil, err
	}
	tlsConfig, err := NewServerTLSConfig(caPem, certPem, keyPem, tls.VerifyClientCertIfGiven)
	if err != nil {
		return nil, err
	}
	pubKey, err := getPublicKeyFromCert(certPem)
	if err != nil {
		return nil, err
	}
	server = &Server{
		serverVersion,
		mysql.MinProtocolVersion,
		capability,
		mysql.DefaultCollationID,
		defaultAuthMethod,
		pubKey,
		tlsConfig,
		new(sync.Map),
		0,
		new(sync.Map),
		util.NewBufferPool(),
		nil,
	}
	return server, nil
}

//Serve on requests
func (server *Server) Serve(port int) error {
	service := fmt.Sprintf(":%d", port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	server.listener = listener
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go server.handle(conn)
	}
}

//Close closes the socket
func (server *Server) Close() {
	if server.listener != nil {
		server.listener.Close()
	}
}

func (server *Server) handle(conn net.Conn) {
	sessionID := atomic.AddUint32(&server.connectionCount, 1)
	s := NewSession(sessionID, server, conn)

	sessions := server.sessions
	sessions.Store(sessionID, s)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovering from panic: %v", r)
		}
		conn.Close()
		sessions.Delete(sessionID)
		s.Close()
	}()

	err := s.AcceptClient()
	if err != nil {
		return
	}
	for {

	}
}
