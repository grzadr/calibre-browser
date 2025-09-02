BIN_DIR = bin

.PHONY: all server

server:
	go build -o $(BIN_DIR)/$@ ./cmd/$@

all: server
