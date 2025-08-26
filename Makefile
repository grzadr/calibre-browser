BIN_DIR = bin

.PHONY: all browser-cli

browser-cli:
	go build -o $(BIN_DIR)/$@ ./cmd/$@

all: browser-cli
