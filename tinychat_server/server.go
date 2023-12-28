package main

import (
	"flag"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const protocol = "tcp"
const host = "localhost"
const port = "8080"

type Status int

const (
	OPEN   Status = 0
	CLOSED        = 1
)

type Session struct {
	Sender   string
	Receiver string
	Channel  chan string
	Status   Status
}

// map to hold active sessions
// each session is identified by a key formed
// in the following way: "<sender>:<receiver>"
var ActiveSessions = make(map[string]Session)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// command line flags
	isDebug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	// logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *isDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Listen for incoming connections
	listener, err := net.Listen(protocol, host+":"+port)
	if err != nil {
		log.Error().Msgf("Error starting the listener: %+v", err.Error())
		return
	}

	defer listener.Close()

	log.Debug().Msgf("Server is listening on port %s", port)

	// Accept incoming connections and handle them
	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Error().Msgf("Error listening to connection: %+v", err.Error())
			continue
		}

		adminCh := make(chan Session)
		defer close(adminCh)

		go HandleInit(conn, adminCh)
		go HandleRead(conn, adminCh)
		go HandleWrite(conn, adminCh)
	}

}
