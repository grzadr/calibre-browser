BIN_DIR = bin

.PHONY: all server client sqlc web

sqlc:
	sqlc generate

server:
	go build -o $(BIN_DIR)/$@ ./cmd/$@

client:
	go build -o $(BIN_DIR)/$@ ./cmd/$@

web:
	go build -o $(BIN_DIR)/$@ .

all: sqlc server client
