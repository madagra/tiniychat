# TinyChat: barebone client-server chat service

TinyChat is a simple Golang application implementing a client-server
chat service based on TCP sockets. This application was mainly developed
for gaining a better understanding of goroutines, channels, and generally
the way Golang handles concurrency. Therefore, no particular attention has
been spent in improving the interface and performance of the code.

See below an example of the application in action.

https://github.com/user-attachments/assets/5bf9c898-9fa6-4244-8197-4f0f3269df07

## Usage

Provided that you have Go installed and available in your path, you can start the
server in a terminal window running `make run_server`.

In separate windows, run the clients using the `make run_client` command. You will
be immediately prompted for your username. After this, you will be able to execute
a series of commands in the chat, which should be prefixed with a `/`, similar
to commonly used messaging systems like Telegram or Slack. The available commands
are:

* `/USERS`: check the full list of chat users. They can be either online or offline.
* `/TIME`: get the current time
* `/START <name>`: start a conversation session with the given user. If the user
never logged into the chat (thus is not in the database), the command will
fail. If the recipient is not currently online, they will receive the messages
immediately after going online.
* `/STOP`: stop the current conversation.
* `/QUIT`: quit the session and notify the server. Keyboard interrupt can also be
used for this purpose.
* `/COMMANDS`: check the list of available commands.

> **DISCLAIMER**: This application has been developed for learning purposes and
there are surely many bugs lurking around.

## Design

The design of the application is pretty basic but makes heavy use of goroutines.

*Client*. The user is immediately prompted at startup with his
username. This username acts as a **unique** identifier for the user. Two
[anonymous](https://www.practical-go-lessons.com/chap-24-anonymous-functions-and-closures)
goroutines in the `main()` function handle *(1)* user prompts,
sending them to the server with the right format and *(2)* messages
from the server, taking care of displaying them with time and sender information
(if in a conversation). A common channel is used to keep track of any error
occurred; the client stops if an error is detected.

> **NOTE**: Clients can start only a single conversation at a time.

*Server*. The server keeps track of the users using a simple in-memory map `Users`
of type `map[string]chan Message`. In this mapping, the keys are the usernames
and the values are pointers `User` datastructures holding information about
the status of the user (online/offline) and a channel used for message delivery.

All messages have a fixed structure defined in the following datastructure:

```go
type Message struct {
    Sender   string
    Receiver string
    Body     string
    Time     string
}
```

JSON serialization is used for sending message via the TCP sockets. The server uses
two goroutines for handling messages:

* `HandleCommands()`: this is the main goroutine which, for each connection (i.e.
online user), listens to the commands sent by the user and process them, sending
back the result.

* `HandleConversation()`: this goroutine deals with conversations. It fetches messages
from the channel associated to each user and send them to the right client.

For communication between the two goroutines, two channels are used: *(1)* a `msgCh`
different for each user (from the `Users` datastructure described above)
listening to all the messages and *(2)* a common `adminCh` used for handling
initialization and termination of the goroutines.

## Possible improvements

The chat is currently very basic. Some possible improvements are the following:

* allow for starting and concluding multiple conversations with the same client session. This can very likely be offloaded to multiple goroutines on the server.
* add basic authentication methods for the users
* allow for sending images and other type of files rather than text.
* move from simple TCP sockets to a more performant gRPC approach using
shared protocol buffer messages for handling conversation and commands.
* use a better CLI library such as [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a more fancy terminal display.
* Add exhaustive unit tests

Since it has been developed mainly for learning purposes, these improvements might
never become a reality.
