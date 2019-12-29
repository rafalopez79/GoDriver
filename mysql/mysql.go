package mysql

import (
	"bytes"
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

//Len returns the lenght
func (packet *Packet) Len() int {
	body := packet.Body
	if body == nil {
		return 0
	}
	return body.Len()
}

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
