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
	"time"
)

type config struct {
	Balencer  []balence
	Prethread int
}

type worker interface {
	run()
	stop()
}

type balence struct {
	Remote   []string
	Local    string
	Protocol string
	stopper  chan bool
}

type pageCache struct {
	Data   []byte
	Host   string
	Proto  string
	Expiry time.Time
}

func main() {
	conf := loadConfigFile()
	for i := 0; i < conf.Prethread; i++ {
		for _, element := range conf.Balencer {
			go element.Run()
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

func (b *balence) Run() {
	b.stopper = make(chan bool)
	go b.run()
}

func (b balence) run() {
	listen, err := net.Listen(b.Protocol, b.Local)
	defer listen.Close()
	acceptChan := make(chan net.Conn)
	perror(err, "Failed to listen on port "+b.Local)
	for i := 0; ; i++ {
		go accept(listen, acceptChan)
		select {
		case _ = <-b.stopper:
			return
		case connect := <-acceptChan:
			go tunnel(connect, b.Remote[i%len(b.Remote)], b.Protocol)
			go accept(listen, acceptChan)
		}

	}
}

func (b balence) Stop() {
	b.stopper <- true
}

func accept(listen net.Listener, conChan chan net.Conn) {
	local, err := listen.Accept()
	perror(err, "failed to accept")

	conChan <- local
}

func tunnel(incoming net.Conn, remoteStr string, proto string) {
	remote, err := net.Dial(proto, remoteStr)
	perror(err, "could not connect to host "+remoteStr)
	go copy(remote, incoming)
	go copy(incoming, remote)
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

func ferror(err error, str string) {
	if err != nil {
		log.Fatalln(str+":", err)
	}
}
