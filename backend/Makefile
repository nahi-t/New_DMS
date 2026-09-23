.PHONY: swagger build run

swagger:
	swag init -g cmd/main.go -o cmd/docs --parseInternal

build: swagger
	go build -o app ./cmd

run: swagger
	go run ./cmd