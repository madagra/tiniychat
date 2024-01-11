package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func HandleConversation(conn net.Conn, adminCh chan string) {

	userName := <-adminCh
	user, _ := Users[userName]
	writer := bufio.NewWriter(conn)
	log.Debug().Msgf("Handling conversations for user %s", userName)

Loop:
	for {
		select {
		case user := <-adminCh:
			log.Debug().Msgf("User %s is now offline, no conversation handling", user)
			break Loop
		case message := <-(*user).msgCh:
			msgJson := Serialize(&message)
			writer.WriteString(msgJson)
			writer.Flush()
		}
	}
}

func HandleCommands(conn net.Conn, adminCh chan string) {

	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// the first string received is always the username
	userName, _ := reader.ReadString('\n')
	userName = strings.ReplaceAll(userName, "\n", "")
	adminCh <- userName

	log.Debug().Msgf("Handling commands for user %s", userName)
	SetUserOnline(userName, conn)

	var inConversation bool = false
	var currentReceiver string = ""

Loop:
	for {

		response, _ := reader.ReadString('\n')

		msgJson := Message{}
		json.Unmarshal([]byte(response), &msgJson)

		var message string
		if len(response) == 0 {
			message = "/" + QUIT
		} else {
			message = msgJson.Body
		}

		cmdOnly := strings.Split(message, " ")[0]
		cmd := Command(strings.TrimSuffix(cmdOnly[1:], "\n"))
		log.Debug().Msgf("User %s received command %s", userName, cmd)

		switch cmd {

		case USERS:
			var users []string
			for key, user := range Users {
				var userMsg string
				if user.isOnline {
					userMsg = fmt.Sprintf("%s <ONLINE>", key)
				} else {
					userMsg = fmt.Sprintf("%s <OFFLINE>", key)
				}
				users = append(users, userMsg)
			}
			msgJson := SerializeFromData("", userName, strings.Join(users, "\n"))
			writer.WriteString(msgJson)
			writer.Flush()

		case COMMANDS:
			cmds := []string{string(USERS), COMMANDS, QUIT, TIME, START, STOP}
			msgJson := SerializeFromData("", userName, strings.Join(cmds, "\n"))
			writer.WriteString(msgJson)
			writer.Flush()

		case TIME:
			currentTime := time.Now().Format("2006-01-02 15:04:05")
			msgJson := SerializeFromData("", userName, currentTime)
			writer.WriteString(msgJson)
			writer.Flush()

		case START:
			receiver := strings.TrimSuffix(strings.Split(message, " ")[1], "\n")
			_, exists := Users[receiver]
			if !exists {
				errMsg := fmt.Sprintf("Cannot start a conversation with user %s. The user does not exist!", strings.TrimSuffix(receiver, "\n"))
				msgJson := SerializeFromData("", userName, errMsg)
				writer.WriteString(msgJson)
				writer.Flush()
			} else {
				inConversation = true
				log.Debug().Msgf("Starting conversation with user %s", receiver)
				currentReceiver = receiver
			}

		case STOP:
			log.Debug().Msgf("Stopping conversation with user %s", currentReceiver)
			inConversation = false
			currentReceiver = ""

		case QUIT:
			inConversation = false
			SetUserOffline(userName)
			adminCh <- currentReceiver
			break Loop

		default:
			if !inConversation {
				errMsg := fmt.Sprintf("Command %s not recognized!", cmd)
				msgJson := SerializeFromData("", userName, errMsg)
				writer.WriteString(msgJson)
			} else {
				user, exists := Users[currentReceiver]
				if !exists {
					log.Error().Msg("Trying to send a message to a non-existing user!")
					continue Loop
				} else {
					user.msgCh <- Message{
						Sender:   userName,
						Receiver: currentReceiver,
						Body:     message,
						Time:     time.Now().Format("2006-01-02 15:04:05"),
					}
				}
			}
			writer.Flush()
		}
	}
}
