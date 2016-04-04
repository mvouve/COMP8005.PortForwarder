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
	Balencer  []balence
	Prethread int
}

// The structure for a load balencer
type balence struct {
	Remote   []string // host to connect to
	Local    string   // local port
	Protocol string   // protocol to use
}

type runner interface {
	run()
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    main
--
-- DATE:        March 7, 2016
--
-- REVISIONS:	  April 1st consolodated forwarder into balencer.
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:		audit(cInfo chan clientInfo)
-- 		 cInfo:		A channel to push client into into.
--
-- RETURNS: 		void
--
-- NOTES:			This function is the main entry point into the program
------------------------------------------------------------------------------*/
func main() {
	conf := loadConfigFile()
	for i := 0; i < conf.Prethread; i++ {
		for _, element := range conf.Balencer {
			go element.run()
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(">")
	reader.ReadString('\n')
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    loadConfigFile
--
-- DATE:        March 7, 2016
--
-- REVISIONS:	  (date and description)
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:		loadConfigFile() config
-- 		 config:	the file configuration as a structure.
--
-- RETURNS: 		void
--
-- NOTES:			This function is a helper function for loading the configuration
--						of the port forwarder
------------------------------------------------------------------------------*/
func loadConfigFile() config {
	var conf config

	j, err := ioutil.ReadFile(os.Args[1])
	ferror(err, "error reading config")
	json.Unmarshal(j, &conf)

	return conf
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    balence run
--
-- DATE:        March 7, 2016
--
-- REVISIONS:	  (date and description)
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:		(b balence) run()
-- 		     b: 	the balencer structure to run.
--
-- RETURNS: 		void
--
-- NOTES:				routines simply sets up tunnels based upon the balencing struct
------------------------------------------------------------------------------*/
func (b balence) run() {
	listen, err := net.Listen(b.Protocol, b.Local)
	perror(err, "Failed to listen on port "+b.Local)
	for i := 0; ; i++ {
		tunnel(listen, b.Remote[i%len(b.Remote)], b.Protocol)
	}
}

/*-----------------------------------------------------------------------------
-- FUNCTION:     tunnel
--
-- DATE:        March 7, 2016
--
-- REVISIONS:	  (date and description)
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:		func tunnel(listen net.Listener, remoteStr string, proto string)
-- 		listen:		A listening server port
-- remoteStr:		The remote host to connect to.
--	   proto:		The protocol to use
--
-- RETURNS: 		void
--
-- NOTES:				routines block until a new connection is made, then creates go
--							routines to tunnel channel full duplex.
------------------------------------------------------------------------------*/
func tunnel(listen net.Listener, remoteStr string, proto string) {
	local, err := listen.Accept()
	perror(err, "failed to accept")
	remote, err := net.Dial(proto, remoteStr)
	perror(err, "could not connect to host "+remoteStr)
	go copy(remote, local)
	go copy(local, remote)
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    copy
--
-- DATE:        March 7, 2016
--
-- REVISIONS:	  (date and description)
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:		copy(src net.Conn, dst net.Conn)
--							src : the source to copy from
--							dst : the destination to copy to.
--
-- RETURNS: 		void
--
-- NOTES:				Helper function to make sure ports are closed when finished
------------------------------------------------------------------------------*/
func copy(src net.Conn, dst net.Conn) {
	defer src.Close()
	defer dst.Close()

	io.Copy(src, dst)
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    perror
--
-- DATE:        March 7, 2016
--
-- REVISIONS:	  (date and description)
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:		func perror(err error, str string)
--			err:		error to check
--			str: 		message to print with error.
--
-- RETURNS: 		void
--
-- NOTES:				Helper function to print error and keep running
------------------------------------------------------------------------------*/
func perror(err error, str string) {
	if err != nil {
		log.Println(str+":", err)
	}
}

/*-----------------------------------------------------------------------------
-- FUNCTION:    ferror
--
-- DATE:        March 7, 2016
--
-- REVISIONS:	  (date and description)
--
-- DESIGNER:		Marc Vouve
--
-- PROGRAMMER:	Marc Vouve
--
-- INTERFACE:		func ferror(err error, str string)
--			err:		error to check
--			str: 		message to print with error.
--
-- RETURNS: 		void
--
-- NOTES:				Helper function to print error and exit
------------------------------------------------------------------------------*/
func ferror(err error, str string) {
	if err != nil {
		log.Fatalln(str+":", err)
	}
}
