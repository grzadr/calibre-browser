BIN_DIR = bin

.PHONY: all server client

server:
	go build -o $(BIN_DIR)/$@ ./cmd/$@

client:
	go build -o $(BIN_DIR)/$@ ./cmd/$@

all: server client
