BIN_DIR = bin

.PHONY: browser-cli

browser-cli:
	go build -o $(BIN_DIR)/$@ ./cmd/$@.go
