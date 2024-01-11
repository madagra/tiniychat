package main

import (
	"encoding/json"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Body     string `json:"body"`
	Time     string `json:"time"`
}

type Command string

const (
	USERS    Command = "USERS"
	COMMANDS         = "COMMANDS"
	QUIT             = "QUIT"
	TIME             = "TIME"
	START            = "START"
	STOP             = "STOP"
)

// maximum number of buffered messages in
// a session channel
const nBufMsg = 1000

// map to hold active users with current connection
var ActiveUsers = make(map[string]net.Conn)

// map to hold the available users who can be either
// online or offline
var Users = make(map[string]chan Message)

func SetUserOnline(userName string, conn net.Conn) chan Message {
	log.Debug().Msgf("Setting user %s online", userName)

	_, exists := Users[userName]
	if !exists {
		Users[userName] = make(chan Message, nBufMsg)
	}

	ActiveUsers[userName] = conn
	return Users[userName]
}

func SetUserOffline(userName string) {
	_, exists := ActiveUsers[userName]
	if !exists {
		log.Warn().Msgf("The user %s cannot be set offline.", userName)
	} else {
		log.Debug().Msgf("Setting user %s offline", userName)
		delete(ActiveUsers, userName)
	}
}

func Serialize(message *Message) string {
	log.Debug().Msgf("Serialize message to %s: %s", message.Sender, message.Body)
	msgJson, err := json.Marshal(message)
	if err != nil {
		log.Error().Msgf("Failed to serialize message: %s", message.Body)
		return "\n"
	}
	return string(msgJson) + "\n"
}

func SerializeFromData(sender string, receiver string, body string) string {
	msg := &Message{
		Sender:   sender,
		Receiver: receiver,
		Body:     body,
		Time:     time.Now().Format("2006-01-02 15:04:05"),
	}
	return Serialize(msg)
}
