BINARY_NAME=app
APP_NAME=tinychat

compile:
	go build -C ${APP_NAME}_server -o `pwd`/${BINARY_NAME}_server
	go build -C ${APP_NAME}_client -o `pwd`/${BINARY_NAME}_client

run_server: compile
	`pwd`/${BINARY_NAME}_server --debug

run_client: compile
	`pwd`/${BINARY_NAME}_client --debug

clean:
	go clean
	rm `pwd`/${BINARY_NAME}_server
	rm `pwd`/${BINARY_NAME}_client
