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

type User struct {
	isOnline bool
	msgCh    chan Message
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
// a session channel. If more that nBufMsg are
// sent while the user is offline, the exceeding ones
// will be discarded
const nBufMsg = 1000

// map to hold the available users
var Users = make(map[string]*User)

func SetUserOnline(userName string, conn net.Conn) *User {
	log.Debug().Msgf("Setting user %s online", userName)

	user, exists := Users[userName]
	if !exists {
		Users[userName] = &User{
			msgCh:    make(chan Message, nBufMsg),
			isOnline: true,
		}
	} else {
		user.isOnline = true
	}
	return Users[userName]
}

func SetUserOffline(userName string) {
	user, exists := Users[userName]
	if !exists {
		log.Warn().Msgf("The user %s cannot be set offline since it does not exist.", userName)
	} else {
		log.Debug().Msgf("Setting user %s offline", userName)
		user.isOnline = false
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
