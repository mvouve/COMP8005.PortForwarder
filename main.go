package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type config struct {
	Forwarder []forward
	Balencer  []balence
}

type forward struct {
	Remote   string
	Local    string
	Protocol string
}

type balence struct {
	Remote   []string
	Local    string
	Protocol string
}

func main() {
	conf := loadConfigFile()
	for _, element := range conf.Balencer {
		go element.run()
	}
	for _, element := range conf.Forwarder {
		go element.run()
	}
	for {
	}
}

func loadConfigFile() config {
	var conf config

	j, err := ioutil.ReadFile(os.Args[1])
	perror(err, "error reading config")
	json.Unmarshal(j, &conf)

	return conf
}

func (f forward) run() {
	listen, err := net.Listen(f.Protocol, f.Local)
	perror(err, "Failed to listen on port "+f.Local)
	for {
		dst, _ := net.Dial(f.Protocol, f.Remote)
		src, _ := listen.Accept()
		go copy(dst, src)
		go copy(src, dst)
	}
}

func (b balence) run() {
	listen, err := net.Listen(b.Protocol, b.Local)
	perror(err, "Failed to listen on port "+b.Local)
	for i := 0; ; i++ {
		local, err := listen.Accept()

		perror(err, "failed to accept")
		remote, err := net.Dial(b.Protocol, b.Remote[i%len(b.Remote)])
		fmt.Println(local.RemoteAddr(), "=>", remote.RemoteAddr())
		perror(err, "could not connect to host "+b.Remote[i%len(b.Remote)])
		go copy(remote, local)
		go copy(local, remote)
	}
}

func copy(src net.Conn, dst net.Conn) {
	defer src.Close()
	defer dst.Close()

	io.Copy(src, dst)
}

func perror(err error, str string) {
	if err != nil {
		log.Println(str+":", err)
	}
}
