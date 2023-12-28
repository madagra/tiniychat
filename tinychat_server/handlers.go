package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const initMessage = "INIT"
const quitMessage = "QUIT"
const waitTime = 2 * time.Second

// maximum number of buffered messages in
// a session channel
const nBufMsg = 100

func closeSession(session Session) Session {

	key := fmt.Sprintf("%s:%s", session.Sender, session.Receiver)
	session.Status = CLOSED

	close(session.Channel)
	delete(ActiveSessions, key)

	log.Debug().Msgf("Removed session from %s to %s with key %s", session.Sender, session.Receiver, key)

	return session
}

// read the first message containing sender and receiver
// and check if the conversation is already ongoing
func HandleInit(conn net.Conn, adminCh chan Session) {

	reader := bufio.NewReader(conn)
	message, _ := reader.ReadString('\n')

	tmp := strings.Split(strings.ReplaceAll(message, "\n", ""), ":")

	var key string
	if tmp[0] == initMessage {

		key = strings.Join(tmp[1:], ":")
		_, exists := ActiveSessions[key]

		if !exists {
			ActiveSessions[key] = Session{
				Sender:   tmp[1],
				Receiver: tmp[2],
				Channel:  make(chan string, nBufMsg),
				Status:   OPEN,
			}
		}

	}

	// send information for both Read and Write routines
	// 3 times because there are 3 calls to the channel at the beginning
	// of the different goroutines
	adminCh <- ActiveSessions[key]
	adminCh <- ActiveSessions[key]
	adminCh <- ActiveSessions[key]
}

// handle reading of new messages coming from
// the input connection. Messages are processed in
// sequence and sent to the channel initialized with
// the given active session
// the `adminCh` variable is just a configuration channel
// which blocks until the handler is ready to listen to
// messages
func HandleRead(conn net.Conn, adminCh chan Session) {

	session := <-adminCh
	reader := bufio.NewReader(conn)
	key := fmt.Sprintf("%s:%s", session.Sender, session.Receiver)

	for {

		message, err := reader.ReadString('\n')
		log.Debug().Msgf("READ - %s - sender: %s - receiver: %s", strings.ReplaceAll(message, "\n", ""), session.Sender, session.Receiver)

		if err != nil {
			log.Error().Msgf("Error %+v detected, connection is not active anymore", err.Error())
			break
		}

		if strings.ReplaceAll(message, "\n", "") == quitMessage {
			break
		}

		// send the message in the channel
		session.Channel <- message

	}

	// notify the admin channel that the session has been closed
	adminCh <- closeSession(ActiveSessions[key])
}

// handle writing of new messages coming from
// the input connection. It uses the sender and receiver
// identifiers of the current session to check if another
// session is available for receiving the messages. If not
// it waits until it is available
// the `adminCh` variable is just a configuration channel
// which blocks until the handler is ready to listen to
// messages
func HandleWrite(conn net.Conn, adminCh chan Session) {

	writer := bufio.NewWriter(conn)

	thisSession := <-adminCh
	otherKey := fmt.Sprintf("%s:%s", thisSession.Receiver, thisSession.Sender)

	for {

		otherSession, otherSessionExists := ActiveSessions[otherKey]
		log.Debug().Msgf("WRITE - Session with key %s exists? %v", otherKey, otherSessionExists)

		if otherSessionExists {
			select {

			case session := <-adminCh:
				if session.Status == CLOSED {
					log.Debug().Msg("WRITE - closing goroutine")
					return
				}

			case message := <-otherSession.Channel:
				_, err := writer.WriteString(message)
				if err != nil {
					log.Error().Msgf("Error %+v detected, connection is not active anymore", err.Error())
					continue
				}
				writer.Flush()

			}

		} else {
			select {

			case session := <-adminCh:
				if session.Status == CLOSED {
					log.Debug().Msg("WRITE - closing goroutine")
					return
				}

			default:
				time.Sleep(waitTime)

			}
		}
	}
}

// just for debugging purposes
func RepeatMessage(conn net.Conn, adminCh chan Session) {

	session := <-adminCh
	writer := bufio.NewWriter(conn)

	for {
		message := <-session.Channel
		_, err := writer.WriteString(message)
		if err != nil {
			log.Error().Msgf("Error %+v detected, connection is not active anymore.\n", err.Error())
			break
		}
		writer.Flush()
	}
}
