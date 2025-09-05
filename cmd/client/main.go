package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/grzadr/calibre-browser/internal/arguments"
	"github.com/grzadr/calibre-browser/internal/booksdb"
)

func run(conf arguments.Config,
	ctx context.Context,
	cancel context.CancelFunc,
) error {
	log.Printf("running server with config:\n%+v\n", conf)
	booksdb.InitializeBooksDb(conf.DbPath, ctx)

	return nil
}

func main() {
	log.Println("running client")
	log.Println("initializing context")
	ctx, cancel := context.WithCancel(context.Background())

	conf, err := arguments.ParseArgsClient(os.Args)
	if err != nil {
		log.Fatalln(fmt.Errorf("error parsing args: %w", err))
	}

	if err := run(conf, ctx, cancel); err != nil {
		log.Fatalln(fmt.Errorf("error running server: %w", err))
	}
}
