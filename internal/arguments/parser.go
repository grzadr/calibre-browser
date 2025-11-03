package arguments

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Config struct {
	DbPath string
}

func validateDbPath(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot access database file '%s': %w", filename, err)
	}

	defer file.Close()

	return nil
}

func ParseArgsServer(args []string) (conf Config, err error) {
	log.Printf("parsing server arguments: %+v", args)

	// Create a new FlagSet for parsing
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s <db filename>\n", args[0])
		fmt.Fprintf(fs.Output(), "\nArguments:\n")
		fmt.Fprintf(fs.Output(), "  db filename    Path to the Calibre database file\n")
	}

	// Parse flags (currently none, but this allows for future additions)
	if err := fs.Parse(args[1:]); err != nil {
		return conf, err
	}

	// Get the positional argument (database path)
	if fs.NArg() < 1 {
		fs.Usage()
		return conf, fmt.Errorf("missing required argument: database path")
	}

	conf.DbPath = fs.Arg(0)

	if err := validateDbPath(conf.DbPath); err != nil {
		return conf, err
	}

	return conf, nil
}
