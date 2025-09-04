package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grzadr/calibre-browser/internal/arguments"
	"github.com/grzadr/calibre-browser/internal/booksdb"
	"github.com/grzadr/calibre-browser/internal/socket"
	_ "modernc.org/sqlite"
)

const (
	defaultErrChanCapacity  = 4
	defaultConnTimeout      = 5 * time.Minute
	defaultConnReadTimeout  = 1 * time.Second
	defaultTerminateTimeout = 10 * time.Second
)

func acceptClientConnection(ctx context.Context,
	listener net.Listener,
	errChan chan<- error,
) chan net.Conn {
	connChan := make(chan net.Conn, 1)

	go func() {
		defer close(connChan)
	loop:
		for {
			conn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Allow shutdown check in next iteration
				}
				select {
				case errChan <- err:
					continue loop
				case <-ctx.Done():
				}
				return
			}

			select {
			case connChan <- conn:
			case <-ctx.Done():
				conn.Close()
				return
			}
		}
	}()
	return connChan
}

func handleClientConnection(
	ctx context.Context,
	conn net.Conn,
	wg *sync.WaitGroup,
	errChan chan<- error,
) {
	wg.Add(1)
	defer wg.Done()
	defer conn.Close()

	// Set connection deadline for graceful shutdown
	conn.SetDeadline(time.Now().Add(defaultConnTimeout))

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	defer writer.Flush()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Handle client communication with timeout
		conn.SetReadDeadline(time.Now().Add(defaultConnReadTimeout))
		message, err := reader.ReadString('\n')
		if err != nil {
			// Check if error is due to shutdown or genuine network issue
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Allow shutdown check in next iteration
			}
			errChan <- err
			return
		}

		response := fmt.Sprintf("Server echo: %s", message)
		writer.WriteString(response)
		writer.Flush()
	}
}

func waitClientConnection(
	ctx context.Context,
	listener net.Listener,
	wg *sync.WaitGroup,
	errChan chan<- error,
) {
	connChan := acceptClientConnection(ctx, listener, errChan)

	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-connChan:
			go handleClientConnection(ctx, conn, wg, errChan)
		}
	}
}

func handleErrors(
	ctx context.Context,
	_ context.CancelFunc,
) chan error {
	errChan := make(chan error, defaultErrChanCapacity)
	go func() {
		// defer close(errChan)
		for {
			select {
			case err := <-errChan:
				log.Println(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return errChan
}

func setupGracefulShutdown(cancel context.CancelFunc, wg *sync.WaitGroup) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating graceful shutdown\n", sig)

		// Cancel context - broadcasts to all goroutines
		cancel()

		// Wait for all client handlers to complete with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			fmt.Println("all client connections closed gracefully")
		case <-time.After(defaultTerminateTimeout):
			log.Println("timeout reached - continuing shutdown")
		}

		log.Println("signal handler completed - returning control to main")
	}()
}

func run(
	conf arguments.Config,
	ctx context.Context,
	cancel context.CancelFunc,
) error {
	log.Printf("running server with config:\n%+v\n", conf)
	booksdb.InitializeBooksDb(conf.DbPath, ctx)

	listener, err := socket.CreateSocketListener(socket.DefaultSocketPath)
	if err != nil {
	}
	defer listener.Close()

	errChan := handleErrors(ctx, cancel)
	defer close(errChan)

	var wg sync.WaitGroup

	setupGracefulShutdown(cancel, &wg)

	go waitClientConnection(ctx, &listener, &wg, errChan)

	<-ctx.Done()
	return nil
}

func main() {
	log.Println("initializing context")
	ctx, cancel := context.WithCancel(context.Background())

	conf, err := arguments.ParseArgsServer(os.Args)
	if err != nil {
		log.Fatalln(fmt.Errorf("error parsing args: %w", err))
	}

	if err := run(conf, ctx, cancel); err != nil {
		log.Fatalln(fmt.Errorf("error running server: %w", err))
	}
}
