package mysql

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"
)

//Packet in the mysql protocol
type Packet struct {
	Body *bytes.Buffer
}

//NewPacket creates a new packet
func NewPacket(options ...func(*Packet)) *Packet {
	packet := Packet{
		Body: nil,
	}
	for _, option := range options {
		option(&packet)
	}
	return &packet
}

//NewPingPacket creates a new Ping Packet
func NewPingPacket() *Packet {
	return NewPacket(func(p *Packet) {
		var data [1]byte
		data[0] = ComPing
		p.Body = bytes.NewBuffer(data[:])
	})
}

//NewSimpleErrPacket creates a new simple Error Packet
func NewSimpleErrPacket(errorCode uint16, msg string) *Packet {
	return NewPacket(func(p *Packet) {
		data := make([]byte, 3+len(msg))
		data[0] = ERRHeader
		//TODO: fields
		p.Body = bytes.NewBuffer(data)
	})
}

//Len returns the lenght
func (packet *Packet) Len() int {
	body := packet.Body
	if body == nil {
		return 0
	}
	return body.Len()
}

//util bytes

//WriteNullTerminatedString append the nts to the buffer
func WriteNullTerminatedString(buffer *bytes.Buffer, s string) (err error) {
	_, err = buffer.WriteString(s)
	if err != nil {
		return err
	}
	return buffer.WriteByte(0)
}

//WriteFixedLengthString append the nts to the buffer
func WriteFixedLengthString(buffer *bytes.Buffer, s string) (err error) {
	_, err = buffer.WriteString(s)
	return err
}

//WriteRLEString append the nts to the buffer
func WriteRLEString(buffer *bytes.Buffer, s string) (err error) {
	err = WriteRLEInt(buffer, uint64(len(s)))
	if err != nil {
		return err
	}
	_, err = buffer.WriteString(s)
	return err
}

//WriteBytes writes byte to buffer
func WriteBytes(buffer *bytes.Buffer, bytes ...byte) (err error) {
	for _, b := range bytes {
		err = buffer.WriteByte(b)
		if err != nil {
			return err
		}
	}
	return nil
}

//Write writes byte to buffer
func Write(buffer *bytes.Buffer, bytes []byte) (err error) {
	n, err := buffer.Write(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return fmt.Errorf("Write failed. only %v bytes written while %v expected", n, len(bytes))
	}
	return nil
}

//WriteInt4 writes int4 to buffer
func WriteInt4(buffer *bytes.Buffer, b uint32) (err error) {
	return WriteBytes(buffer, byte(b), byte(b>>8), byte(b>>16), byte(b>>24))
}

//WriteInt3 writes int3 to buffer
func WriteInt3(buffer *bytes.Buffer, b uint32) (err error) {
	return WriteBytes(buffer, byte(b), byte(b>>8), byte(b>>16))
}

//WriteInt2 writes int2 to buffer
func WriteInt2(buffer *bytes.Buffer, b uint16) (err error) {
	return WriteBytes(buffer, byte(b), byte(b>>8))
}

//Read read byte from buffer
func Read(buffer *bytes.Buffer, bytes []byte) (err error) {
	n, err := buffer.Read(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return fmt.Errorf("Read failed. only %v bytes read while %v expected", n, len(bytes))
	}
	return nil
}

//ReadInt4 writes int4 to buffer
func ReadInt4(buffer *bytes.Buffer) (b uint32, err error) {
	var data [4]byte
	err = Read(buffer, data[:])
	if err != nil {
		return 0, err
	}
	b = uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24
	return b, nil
}

//ReadInt3 reads int3 to buffer
func ReadInt3(buffer *bytes.Buffer) (b uint32, err error) {
	var data [3]byte
	err = Read(buffer, data[:])
	if err != nil {
		return 0, err
	}
	b = uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16
	return b, nil
}

//ReadInt2 reads int2 to buffer
func ReadInt2(buffer *bytes.Buffer) (b uint16, err error) {
	var data [2]byte
	err = Read(buffer, data[:])
	if err != nil {
		return 0, err
	}
	b = uint16(data[0]) + uint16(data[1])<<8
	return b, nil
}

//WriteRLEInt writes int to buffer
func WriteRLEInt(buffer *bytes.Buffer, n uint64) (err error) {
	switch {
	case n <= 250:
		return buffer.WriteByte(byte(n))
	case n <= 0xffff:
		b0 := byte(n)
		b1 := byte(n >> 8)
		return WriteBytes(buffer, twoByte, b0, b1)
	case n <= 0xffffff:
		b0 := byte(n)
		b1 := byte(n >> 8)
		b2 := byte(n >> 16)
		return WriteBytes(buffer, threeByte, b0, b1, b2)
	case n <= 0xffffffffffffffff:
		b0 := byte(n)
		b1 := byte(n >> 8)
		b2 := byte(n >> 16)
		b3 := byte(n >> 24)
		b4 := byte(n >> 32)
		b5 := byte(n >> 40)
		b6 := byte(n >> 48)
		b7 := byte(n >> 56)
		return WriteBytes(buffer, eightByte, b0, b1, b2, b3, b4, b5, b6, b7)
	}
	return fmt.Errorf("Wrong uint64: %v", n)
}

//WriteRLEIntNUL writes bull int to buffer
func WriteRLEIntNUL(buffer *bytes.Buffer) error {
	return buffer.WriteByte(oneByte)
}

//ReadRLEInt reads int to buffer
func ReadRLEInt(buffer *bytes.Buffer) (num int64, null bool, err error) {
	var mark byte
	mark, err = buffer.ReadByte()
	if mark < oneByte {
		num := int64(mark)
		return num, false, err
	} else if mark == oneByte {
		return 0, true, nil
	} else if mark == twoByte {
		b1, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b2, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		return int64(b1) + int64(b2)<<8, false, nil
	} else if mark == threeByte {
		b1, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b2, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b3, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		return int64(b1) + int64(b2)<<8 + int64(b3)<<16, false, nil
	} else if mark == eightByte {
		b1, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b2, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b3, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b4, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b5, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b6, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b7, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		return int64(b1) + int64(b2)<<8 + int64(b3)<<16 + int64(b4)<<24 + int64(b4)<<32 + int64(b5)<<40 + int64(b6)<<48 + int64(b7)<<56, false, nil
	}
	return 0, false, fmt.Errorf("Wrong RLE Integer")
}

//ReadFixedLengthString reads string from byte[]
func ReadFixedLengthString(buff []byte, l int) string {
	return String(buff[:Min(len(buff), l)])
}

//Min of 2 ints
func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

//Max of 2 ints
func Max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

//String from slice
func String(b []byte) (s string) {
	bytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	data := (*reflect.StringHeader)(unsafe.Pointer(&s))
	data.Data = bytes.Data
	data.Len = bytes.Len
	return
}

//packed integer
const (
	oneByte   byte = 0XFB
	twoByte   byte = 0XFC
	threeByte byte = 0XFD
	eightByte byte = 0XFE
)

//MaxPayloadLen of packet
const (
	MaxPayloadLen      int  = 1<<24 - 1
	MinProtocolVersion byte = 10
)

//COMMANDS
const (
	ComSleep = iota
	ComQuit
	ComInitDB
	ComQuery
	ComFieldList
	ComCreateDB
	ComDropDB
	ComRefresh
	ComShutdown
	ComStatistics
	ComProcessInfo
	ComConnect
	ComProcessKill
	ComDebug
	ComPing
	ComTime
	ComDelayedInsert
	ComChangeUser
	ComBinlogDump
	ComTableDump
	ComConnectoOut
	ComRegisterSlave
	ComSTMTPrepare
	ComSTMTExecute
	ComSTMTSendLongData
	ComSTMTClose
	ComSTMTReset
	ComSetOption
	ComSTMTFetch
	ComDaemon
	ComUnimplemented
	ComResetConnection
)

//Server
const (
	ServerStatusInTrans           uint16 = 0x0001
	ServerStatusAutocommit        uint16 = 0x0002
	ServerStatusMoreResultsExists uint16 = 0x0008
	ServerStatusNoGoodIndexUsed   uint16 = 0x0010
	ServerStatusNoIndexUsed       uint16 = 0x0020
	ServerStatusCursorExists      uint16 = 0x0040
	ServerStatusLastRowSend       uint16 = 0x0080
	ServerStatusDBDroppped        uint16 = 0x0100
	ServerStatusNoBackslashScaped uint16 = 0x0200
	ServerStatusMetadataChanged   uint16 = 0x0400
	ServerStatusQueryWasLow       uint16 = 0x0800
	ServerStatusPSOutParams       uint16 = 0x1000
)

//STMT Indicator
const (
	STMTIndicatorNone = iota
	STMTIndicatorNull
	STMTIndicatorDefault
	STMTIndicatorIgnore
)

//AUTH
const (
	AuthMYSQLOldPassword    = "mysql_old_password"
	AuthNativePassword      = "mysql_native_password"
	AuthCachingSHA2Password = "caching_sha2_password"
	AuthSHA2Password        = "sha256_password"
)

//CHARSET
const (
	DefaultCharset             = "utf8"
	DefaultCollationID   uint8 = 33
	DefaultCollationName       = "utf8_general_ci"
)

//HEADER
const (
	OKHeader          byte = 0x00
	MoreDataHeader    byte = 0x01
	ERRHeader         byte = 0xff
	EOFHeader         byte = 0xfe
	LocalInfileHeader byte = 0xfb

	CacheSHA2FastAuth byte = 0x03
	CacheSHA2FullAuth byte = 0x04
)

//Client
const (
	ClientLongPassword uint32 = 1 << iota
	ClientFoundRows
	ClientLongFlag
	ClientConnectWithDB
	ClientNoSchema
	ClientCompress
	ClientODBC
	ClientLocalFiles
	ClientIgnoreSpace
	ClientProtocol41
	ClientInteractive
	ClientSSL
	ClientIgnoreSIGPIPE
	ClientTransactions
	ClientReserved
	ClientSecureConnection
	ClientMultiStatements
	ClientMultiResults
	ClientPSMultiResults
	ClientPluginAuth
	ClientConnectATTRS
	ClientPluginAuthLENENCClientData
)

//MYSQLTYPE
const (
	MYSQLTypeDecimal byte = iota
	MYSQLTypeTiny
	MYSQLTypeShort
	MYSQLTypeLong
	MYSQLTypeFloat
	MYSQLTypeDouble
	MYSQLTypeNull
	MYSQLTypeTimestamp
	MYSQLTypeLongLong
	MYSQLTypeInt24
	MYSQLTypeDate
	MYSQLTypeTime
	MYSQLTypeDateTime
	MYSQLTypeYear
	MYSQLTypeNewDate
	MYSQLtypeVarchar
	MYSQLTypeBit

	//mysql 5.6
	MYSQLTypeTimestamp2
	MYSQLTypeDateTime2
	MYSQLTypeTime2
)

//MYSQL Types
const (
	MYSQLTypeJSON byte = iota + 0xf5
	MYSQLTypeNewDecimal
	MYSQLTypeEnum
	MYSQLTypeSet
	MYSQLTypeTinyBlob
	MYSQLTypeMediumBlob
	MYSQLTypeLongBlob
	MYSQLTypeBlob
	MYSQLTypeVarString
	MYSQLTypeString
	MYSQLTypeGeometry
)

//FLAGS
const (
	NotNullFlag       = 1
	PriKeyFlag        = 2
	UniqueKeyFlag     = 4
	BlobFlag          = 16
	UnsignedFlag      = 32
	ZerofillFlag      = 64
	BinaryFlag        = 128
	EnumFlag          = 256
	AutoIncrementFlag = 512
	TimestampFlag     = 1024
	SetFlag           = 2048
	NumFlag           = 32768
	PartKeyFlag       = 16384
	GoupFlag          = 32768
	UniqueFlag        = 65536
)
