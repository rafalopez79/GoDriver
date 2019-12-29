package server

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	util "github.com/rafalopez79/godriver/internal/util"
	mysql "github.com/rafalopez79/godriver/mysql"
)

//Session in server side
type Session struct {
	sessionID     uint32
	server        *Server
	conn          net.Conn
	writer        io.Writer
	reader        io.Reader
	seq           byte
	bufferPool    *util.BufferPool
	salt          []byte // 8 + 12
	rand          *rand.Rand
	capability    uint32
	maxPacketSize uint32
	collation     byte
}

//NewSession creates a new session
func NewSession(sessionID uint32, server *Server, conn net.Conn) *Session {
	var reader io.Reader = conn
	var writer io.Writer = conn
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	salt, _ := util.RandomBuffer(rand, 20)
	return &Session{
		sessionID,
		server,
		conn,
		writer,
		reader,
		0,
		server.bufferPool,
		salt,
		rand,
		0,
		0,
		0,
	}
}

//AcceptClient performs the connect phase
func (session *Session) AcceptClient() (err error) {
	err = session.writeInitialHandShakePacket()
	if err != nil {
		return err
	}
	var useSSL bool
	useSSL, err = session.readClientHandShakePacket()
	if err != nil {
		return err
	}
	if useSSL {
		//switch to tls
		tlsConn := tls.Server(session.conn, session.server.tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			return err
		}
		session.conn = tlsConn
		session.reader = tlsConn
		session.writer = tlsConn
		useSSL, err = session.readClientHandShakePacket()
		if err != nil {
			return err
		}
	} else {

	}
	return nil
}

//Handle client request after client accept
func (session *Session) Handle() (err error) {
	//server := session.server

	//read packet
	//packetIn, err := session.readPacket()
	//write response

	return nil
}

//Close closes session related resources
func (session *Session) Close() error {
	//TODO release resources
	return nil
}

func isAuthMethodSupported(authMethod string) bool {
	return authMethod == mysql.AuthNativePassword ||
		authMethod == mysql.AuthCachingSHA2Password ||
		authMethod == mysql.AuthSHA2Password
}

func (session *Session) writeInitialHandShakePacket() (err error) {
	buffer := session.bufferPool.Get()
	defer session.bufferPool.Return(buffer)

	server := session.server
	//proto version
	mysql.WriteBytes(buffer, server.protocolVersion)
	//server version
	mysql.WriteNullTerminatedString(buffer, server.serverVersion)
	//conn id
	mysql.WriteInt4(buffer, session.sessionID)
	//salt
	mysql.Write(buffer, session.salt[:8])
	mysql.WriteBytes(buffer, 0)
	//server caps

	//server default collation
	mysql.WriteBytes(buffer, server.collationID)
	//status flags

	//server caps 2

	//if
	mysql.WriteBytes(buffer, 0, 0, 0, 0, 0, 0)
	//if

	//if

	packet := mysql.NewPacket(func(p *mysql.Packet) {
		p.Body = buffer
	})
	return session.writePacket(packet)
}

func (session *Session) readClientHandShakePacket() (useSSL bool, err error) {
	packet, err := session.readSimplePacket(session.reader)
	if err != nil {
		return false, err
	}
	//check SSL req
	len := packet.Len()
	useSSL = len == 32 || len == 5
	server := session.server
	if useSSL {
		if server.capability&mysql.ClientSSL == 0 {
			return false, fmt.Errorf("SSL not supported by server")
		}
	}
	if len == 5 && server.capability&mysql.ClientProtocol41 != 0 {
		//protocol41
		capability, err := mysql.ReadInt2(packet.Body)
		if err != nil {
			return false, err
		}
		session.capability = uint32(capability)
		maxPacketSize, err := mysql.ReadInt3(packet.Body)
		if err != nil {
			return false, err
		}
		session.maxPacketSize = maxPacketSize
		return true, nil
	} else if len < 32 {
		return false, fmt.Errorf("Wrong client handshake packet lenght")
	}
	capability, err := mysql.ReadInt4(packet.Body)
	if err != nil {
		return false, err
	}
	session.capability = capability
	maxPacketSize, err := mysql.ReadInt4(packet.Body)
	if err != nil {
		return false, err
	}
	session.maxPacketSize = maxPacketSize
	collation, err := packet.Body.ReadByte()
	if err != nil {
		return false, err
	}
	session.collation = collation
	packet.Body.Next(19) //reserved
	var clientCapsExtra uint32
	if session.server.capability&mysql.ClientProtocol41 != 0 {
		clientCapsExtra, err = mysql.ReadInt4(packet.Body)
		if err != nil {
			return false, err
		}
		//TODO
		clientCapsExtra++
	} else {
		clientCapsExtra = 0
	}
	if len > 32 {
		//not a ssl req, nrmal handshake

	}
	return useSSL, nil
}

func (session *Session) readPacket() (packets []mysql.Packet, err error) {

	return nil, nil
}

func (session *Session) writePacket(p *mysql.Packet) (err error) {
	const max = mysql.MaxPayloadLen
	var header [4]byte

	writer := session.writer
	body := p.Body.Bytes()
	len := len(body)

	for len >= max {
		header[0] = byte(0xff)
		header[1] = byte(0xff)
		header[2] = byte(0xff)
		header[3] = byte(session.seq)
		err = write(writer, header[:])
		if err != nil {
			return err
		}
		err = write(writer, body[:max])
		if err != nil {
			return err
		}
		session.seq++
		len -= max
		body = body[max:]
	}
	header[0] = byte(len)
	header[1] = byte(len >> 8)
	header[2] = byte(len >> 16)
	header[3] = byte(session.seq)
	err = write(writer, header[:])
	if err != nil {
		return err
	}
	return write(writer, body)
}

func write(writer io.Writer, buff []byte) (err error) {
	n, err := writer.Write(buff)
	if err != nil {
		return err
	} else if n != len(buff) {
		return fmt.Errorf("Write failed. only %v bytes written while %v expected", n, len(buff))
	}
	return nil
}

func (session *Session) resetSeq() {
	session.seq = 0
}

//readSimplePacket reads a packet
func (session *Session) readSimplePacket(reader io.Reader) (packet *mysql.Packet, err error) {
	var n int
	var header [4]byte
	n, err = reader.Read(header[:])
	if err != nil {
		return nil, err
	} else if n != 4 {
		return nil, fmt.Errorf("Wrong byte number read %d", n)
	}
	var len int = int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	var seq byte = header[3]
	if seq != session.seq {
		return nil, fmt.Errorf("invalid sequence %d != %d", seq, session.seq)
	}
	session.seq++
	buff := make([]byte, len)
	n, err = reader.Read(buff)
	if err != nil {
		return nil, err
	} else if n != len {
		return nil, fmt.Errorf("Read failed. only %d bytes read while %d expected", n, len)
	}
	packet = mysql.NewPacket(func(p *mysql.Packet) {
		p.Body = bytes.NewBuffer(buff)
	})
	return packet, nil
}
