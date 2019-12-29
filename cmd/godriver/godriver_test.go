package main

import (
	"testing"

	server "github.com/rafalopez79/godriver/internal/server"
	"github.com/rafalopez79/godriver/mysql"
	"gotest.tools/assert"
)

func TestLauncher(t *testing.T) {
	urlConfig := "../../docs/config.json"
	config, err := getConfig(urlConfig)
	assert.NilError(t, err, "Err must be nil")
	_, err = server.NewServer(config.ServerVersion, mysql.AuthNativePassword)
	assert.NilError(t, err, "Err must be nil")
	//ch := make(chan error)
	//go func() {
	//	err = s.Serve(config.ServerPort)
	//		ch <- err
	//	}()
	//	time.Sleep(1 * time.Second)
	//	s.Close()
	//	err = <-ch
	//	assert.Error(t, err, "Err must be nil")
}
