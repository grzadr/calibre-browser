package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/grzadr/calibre-browser/internal/model"
	_ "modernc.org/sqlite"
)

type Config struct {
	dbPath string
}

const (
	requiredArgsSize = 1
	argsDbPath       = 0
)

func usage(cmd string) string {
	return fmt.Sprintf("%s <db filename>", cmd)
}

func validateDbPath(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot access database file '%s': %w", filename, err)
	}
	defer file.Close()
	return nil
}

func ParseArgs(args []string) (conf Config, err error) {
	cmd := args[0]
	args = args[1:]
	if len(args) < requiredArgsSize {
		return conf, fmt.Errorf(
			"required %d arguments\n%s",
			requiredArgsSize,
			usage(cmd),
		)
	}

	conf.dbPath = args[argsDbPath]

	err = validateDbPath(conf.dbPath)

	return
}

func run(args []string, ctx context.Context) error {
	conf, err := ParseArgs(args)
	if err != nil {
		return fmt.Errorf("error parsing args: %w", err)
	}

	db, err := sql.Open("sqlite", conf.dbPath)
	if err != nil {
		return fmt.Errorf("error opening db %q: %w", conf.dbPath, err)
	}


	queries := model.New(db)

	books, err := queries.ListAllBooks(ctx)

	for _, b := range books {
		log.Println(b)
	}
	return nil
}

func main() {
	ctx := context.Background()

	if err := run(os.Args, ctx); err != nil {
		log.Fatalln(err)
	}
}
