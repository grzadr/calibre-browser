package socket

import (
	"fmt"
	"net"
	"os"
)

const DefaultSocketPath = "/tmp/booksdb.sock"

func CreateSocketListener() (listener net.Listener, err error) {
	// Ensure clean socket file state
    os.Remove(DefaultSocketPath)

    listener, err = net.Listen("unix", DefaultSocketPath)
    if err != nil {
        panic(fmt.Sprintf("Failed to create Unix socket listener: %v", err))
    }

	return
}
