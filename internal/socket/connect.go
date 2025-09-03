package socket

import (
	"fmt"
	"net"
	"os"
)

const (
	DefaultSocketPath = "/tmp/booksdb.sock"
	readOnlyPerm      = 0o600 // rw-------
)

type SocketListener struct {
	socketPath string
	net.Listener
}

func (l *SocketListener) Close() error {
	if err := l.Listener.Close(); err != nil {
		return fmt.Errorf("error closing listener %q: %v\n", l.socketPath, err)
	}

	// Remove socket file - prevents filesystem pollution
	if err := os.Remove(l.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing socket %q: %v\n", l.socketPath, err)
	}

	return nil
}

func CreateSocketListener(
	socketPath string,
) (listener SocketListener, err error) {
	// Ensure clean socket file state
	os.Remove(socketPath)
	listener.socketPath = socketPath

	listener.Listener, err = net.Listen("unix", socketPath)
	if err != nil {
		return listener, fmt.Errorf(
			"failed to create %q socket listener: %w",
			socketPath,
			err,
		)
	}

	if err = os.Chmod(socketPath, readOnlyPerm); err != nil {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		return listener, fmt.Errorf(
			"failed to set socket %q permissions: %w",
			socketPath,
			err,
		)
	}

	return listener, nil
}
