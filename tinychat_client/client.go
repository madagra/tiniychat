package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const protocol = "tcp"
const host = "localhost"
const port = "8080"

func getUserInput(msg string, reader *bufio.Reader) (string, error) {

	// Prompt the user for input
	fmt.Printf("%s", msg)

	// Read user input from the terminal
	input, err := reader.ReadString('\n')
	return input, err
}

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

	// Connect to the server
	conn, err := net.Dial(protocol, host+":"+port)
	if err != nil {
		log.Error().Msgf("Cannot connect to the server: %+v", err)
		return
	}

	defer conn.Close()

	// define the readers
	readerStdin := bufio.NewReader(os.Stdin)
	readerConn := bufio.NewReader(conn)
	writerConn := bufio.NewWriter(conn)

	// send out any errors happening in reading or writing
	// and close the connection if this happens
	errCh := make(chan error)

	// get the name of the user to initiate the
	// session on the server
	userName, _ := getUserInput("Enter your name: ", readerStdin)
	_, err = writerConn.WriteString(userName)
	if err != nil {
		errCh <- err
	}
	writerConn.Flush()

	// handle user input in a separate goroutine (anonymous)
	go func() {

		for {
			message, errRead := getUserInput("", readerStdin)
			if errRead != nil {
				errCh <- errRead
				break
			}

			data := map[string]string{
				"sender": userName,
				"body":   message,
			}
			dataJson, _ := json.Marshal(data)

			_, errWrite := writerConn.WriteString(string(dataJson) + "\n")
			if errWrite != nil {
				errCh <- errWrite
				break
			}

			writerConn.Flush()
			log.Debug().Msgf("Message wrote: %s", message)
		}

	}()

	// handle messages received in a separate goroutine (anonymous)
	go func() {
		for {
			response, err := readerConn.ReadString('\n')
			log.Debug().Msgf("Message received: %s", response)
			if err != nil {
				errCh <- err
				break
			}

			var data map[string]interface{}
			err = json.Unmarshal([]byte(response), &data)
			if err != nil {
				log.Error().Msgf("Failed to deserialize message: %s", response)
				continue
			}
			var sender string = data["sender"].(string)
			var body string = data["body"].(string)
			var time string = data["time"].(string)

			if sender != "" {
				fmt.Printf("<%s,%s> %s\n", sender, time, strings.TrimSuffix(body, "\n"))
			} else {
				fmt.Printf("<> %s\n", strings.TrimSuffix(body, "\n"))
			}
		}
	}()

	// listen for any error and return if one is found
	err = <-errCh
	log.Error().Msgf("Error %+v detected. Quitting the client", err)
	return
}
