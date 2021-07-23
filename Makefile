docker:
	GOOS=linux GOARCH=amd64 go build -o output/client client/client.go
	GOOS=linux GOARCH=amd64 go build -o output/server server/server.go