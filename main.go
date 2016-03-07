package main

import (
	"bufio"
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
	Prethread int
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

type runner interface {
	run()
}

func main() {
	conf := loadConfigFile()
	for i := 0; i < conf.Prethread; i++ {
		for _, element := range conf.Balencer {
			go element.run()
		}
		for _, element := range conf.Forwarder {
			go element.run()
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(">")
	reader.ReadString('\n')
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
		tunnel(listen, f.Remote, f.Protocol)
	}
}

func (b balence) run() {
	listen, err := net.Listen(b.Protocol, b.Local)
	perror(err, "Failed to listen on port "+b.Local)
	for i := 0; ; i++ {
		tunnel(listen, b.Remote[i%len(b.Remote)], b.Protocol)
	}
}

func tunnel(listen net.Listener, remoteStr string, proto string) {
	local, err := listen.Accept()
	perror(err, "failed to accept")
	remote, err := net.Dial(proto, remoteStr)
	perror(err, "could not connect to host "+remoteStr)
	go copy(remote, local)
	go copy(local, remote)
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
