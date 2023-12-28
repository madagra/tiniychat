package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const protocol = "tcp"
const host = "localhost"
const port = "8080"

const initMessage = "INIT"
const quitMessage = "QUIT"

func getUserInput(msg string, reader *bufio.Reader) string {

	// Prompt the user for input
	fmt.Printf("%s", msg)

	// Read user input from the terminal
	userInput, err := reader.ReadString('\n')
	if err != nil {
		panic(fmt.Sprint("Error reading input: ", err))
	}

	return userInput
}

func sendMsg(msg string, conn net.Conn, writer *bufio.Writer) {
	_, err := writer.WriteString(msg)
	if err != nil {
		panic(fmt.Sprint("Error sending message to server: ", err))
	}
	writer.Flush()
}

func handleWrite(conn net.Conn) {

	reader := bufio.NewReader(conn)

	for {

		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection is not available anymore: ", err.Error())
			break
		}
		fmt.Printf("> %s", message)
	}
}

func main() {

	reader := bufio.NewReader(os.Stdin)

	userName := getUserInput("Enter your name: ", reader)
	receiverName := getUserInput("Enter the person you want to chat with: ", reader)

	// Connect to the server
	conn, err := net.Dial(protocol, host+":"+port)
	if err != nil {
		fmt.Println(err)
		return
	}
	writer := bufio.NewWriter(conn)

	defer conn.Close()

	// Send some data to the server
	senderMessage := strings.ReplaceAll(fmt.Sprintf("%s", userName), "\n", "")
	receiverMessage := strings.ReplaceAll(fmt.Sprintf("%s", receiverName), "\n", "")
	sendMsg(fmt.Sprintf("%s:%s:%s\n", initMessage, senderMessage, receiverMessage), conn, writer)

	go handleWrite(conn)
	for {
		message := getUserInput("", reader)
		sendMsg(message, conn, writer)

		if strings.ReplaceAll(message, "\n", "") == quitMessage {
			break
		}
	}

}
