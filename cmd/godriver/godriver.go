package main

import (
	"bufio"
	"flag"
	ioutil "io/ioutil"
	"log"
	"os"

	config "github.com/rafalopez79/godriver/internal/config"
	server "github.com/rafalopez79/godriver/internal/server"
	util "github.com/rafalopez79/godriver/internal/util"
	"github.com/rafalopez79/godriver/mysql"
)

func initLog() *bufio.Writer {
	out := bufio.NewWriterSize(os.Stdout, 64*1024)
	log.SetOutput(out)
	log.SetPrefix("")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	return out
}

func parseArgs() string {
	var configurl string
	flag.StringVar(&configurl, "config", "", "configuration url")
	flag.Parse()
	if configurl == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return configurl
}

func getConfig(url string) (*config.Configuration, error) {
	read, err := util.OpenURI(url)
	if err != nil {
		return nil, err
	}
	defer read.Close()
	data, err := ioutil.ReadAll(read)
	if err != nil {
		return nil, err
	}
	return config.Parse(data)
}

// main function
func main() {

	configurl := parseArgs()
	logout := initLog()
	defer logout.Flush()

	log.Print(configurl)
	log.Println("Starting godriver")
	config, err := getConfig(configurl)
	if err != nil {
		log.Panic("Config not valid", err)
	}
	s, err := server.NewServer(config.ServerVersion, mysql.AuthNativePassword)
	if err != nil {
		log.Panic("Error creating server", err)
	}
	s.Serve(config.ServerPort)
	//close
}
