# TinyChat

TinyChat is an extremely simple application written in Golang implementing
a chat service using TCP sockets. The main purpose of this application was to
experiment with goroutines, channels and concurrency in general using Golang.

## Usage

Provided that you have Go installed and available in your path, you can start the
server in a terminal window running `make run_server`.

In separate windows, run the clients using the `make run_client` command. You will
be immediately prompted for your username and the name of the person you want to
chat with and can start the conversation if the user is active or just send the messages
until the user becomes active. For quitting a session, just use keyboard interrupt or
type "QUIT".

## Design

The design of the application is very basic. Each client is prompted, at startup,
with his name and the name of the person they want to chat with. These names are
unique identifier of the session since they indicate sender and receiver. Two
goroutines handle reading (from the TCP connection) and writing (to the receiver
TCP connection) of the messages. A channel stored in the session datastructure is
used to communicate the message and synchronize the routines.

The active sessions are stored in a map on the server using the following 
datastructure:

```go
type Session struct {
    Sender   string
    Receiver string
    Channel  chan string
    Status   Status
}
```

The session mapping has keys of the following format: `"<sender>:<receiver>"`. For example,
if John starts a session with Marie, the key will be `"John:Marie"`. All the messages directed
to Marie from John will be received and written on the socket corresponding to the session with
key `"Marie:John"`. If this session is not active, the goroutine will simply wait until it is and
cache the messages.

## Improvements

The chat is currently very basic. Some possible improvements:

* add the possibility to execute commands (similarly to, e.g., Slack) such as `/ONLINE` for getting 
the people online, `/TALK` to start a conversation, `/QUIT` to close the session etc.
* allow for starting and concluding multiple conversations with the same client session. This can very
likely be offloaded to multiple goroutines on the server.
* allow for sending images and other type of files rather than text.
