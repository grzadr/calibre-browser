package main

import (
	"context"
	"encoding/json"
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
	log.Printf("running client with config:\n%+v\n", conf)

	if err := booksdb.PopulateBooksRepository(conf.DbPath, ctx); err != nil {
		return fmt.Errorf("error initializng database %q: %w", conf.DbPath, err)
	}

	entries, err := booksdb.ExecuteCommand(
		booksdb.GetBooksEntries(),
		conf.Cmd,
		conf.Args,
	)
	if err != nil {
		return fmt.Errorf("error running command %q: %w", conf.Cmd, err)
	}

	json, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling entries %q: %w", conf.Cmd, err)
	}

	os.Stdout.Write(json)

	log.Println("completed client run")

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
