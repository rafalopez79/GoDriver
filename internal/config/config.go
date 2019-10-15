package config

import (
	"encoding/json"
)

//Connection cloud info
type Connection struct {
	ID       string `json:"id"  binding:"required"`
	User     string `json:"user"  binding:"required"`
	Password string `json:"password"  binding:"required"`
	DBName   string `json:"dbname"  binding:"required"`
}

//Configuration server config
type Configuration struct {
	ServerVersion   string       `json:"serverversion" binding:"required"`
	ProtocolVersion int          `json:"protocolversion" binding:"required"`
	ServerPort      int          `json:"serverport" binding:"required"`
	WebPort         int          `json:"webport" binding:"required"`
	Connections     []Connection `json:"connections" binding:"required"`
}

//Parse the string
func Parse(txt []byte) (*Configuration, error) {
	var c Configuration
	err := json.Unmarshal(txt, &c)
	return &c, err
}
