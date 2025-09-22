BIN_DIR = bin

.PHONY: all server client sqlc

sqlc:
	sqlc generate

server:
	go build -o $(BIN_DIR)/$@ ./cmd/$@

client:
	go build -o $(BIN_DIR)/$@ ./cmd/$@

all: sqlc server client
