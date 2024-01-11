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

/*
Protocol for chatting:

Standard protocol
1. client initiate connection
2. server acknowledge and add the user to the list of online users
3. client can send a series of commands such as /USERS, /ONLINE, /QUIT, /TIME or /START or /STOP
4. server answer to the client with the result of the command

Conversation protocol
1. the client send a /START command to the server to start a conversation
2. the server acknowledge and expect a receiver in the following up command containing the name of the user to talk to
3. the client sends the receiver information and it is then ready to send messages
4. if the receiver is online, the message is delivered directly by the server
5. if the receiver is offline, the message is buffered and sent as soon as the receiver comes online
6. the client sends a /STOP command to tell the server that the conversation ends
*/

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

		adminCh := make(chan string)

		defer conn.Close()
		defer close(adminCh)

		go HandleCommands(conn, adminCh)
		go HandleConversation(conn, adminCh)
	}

}
