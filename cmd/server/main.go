package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/grzadr/calibre-browser/internal/arguments"
	"github.com/grzadr/calibre-browser/internal/booksdb"
	"github.com/grzadr/calibre-browser/internal/socket"
	_ "modernc.org/sqlite"
)



func run(conf arguments.Config, ctx context.Context) error {
	booksdb.InitializeBooksDb(conf.DbPath, ctx)

	listener, err := socket.CreateSocketListener()

	if err != nil {

	}


	return nil
}

func main() {
	ctx := context.Background()

	conf, err := arguments.ParseArgs(os.Args)
	if err != nil {
		log.Fatalln(fmt.Errorf("error parsing args: %w", err))
	}

	if err := run(conf, ctx); err != nil {
		log.Fatalln(fmt.Errorf("error running server: %w", "err"))
	}
}
