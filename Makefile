BINARY_NAME=app
APP_NAME=tinychat

compile:
	go build -C ${APP_NAME}_server -o `pwd`/${BINARY_NAME}_server

run_server: compile
	`pwd`/${BINARY_NAME}_server --debug

run_client:
	go run `pwd`/${APP_NAME}_client/client.go

clean:
	go clean
	rm `pwd`/${BINARY_NAME}_server
