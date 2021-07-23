docker:
	GOOS=linux GOARCH=amd64 go build client/client.go
	GOOS=linux GOARCH=amd64 go build server/server.go