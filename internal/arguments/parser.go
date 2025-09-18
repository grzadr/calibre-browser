package arguments

import (
	"fmt"
	"log"
	"os"
)

const (
	requiredArgsSize = 1
	argsDbPath       = 0
	argsCmd          = 1
)

type Config struct {
	DbPath string
	Cmd    string
	Args   []string
}

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

func ParseArgsServer(args []string) (conf Config, err error) {
	log.Printf("parsing server arguments: %+v", args)

	cmd := args[0]
	args = args[1:]

	if len(args) < requiredArgsSize {
		return conf, fmt.Errorf(
			"required %d arguments\n%s",
			requiredArgsSize,
			usage(cmd),
		)
	}

	conf.DbPath = args[argsDbPath]
	conf.Args = args[requiredArgsSize+1:]

	err = validateDbPath(conf.DbPath)

	return
}

func ParseArgsClient(args []string) (conf Config, err error) {
	log.Printf("parsing client arguments: %+v", args)

	cmd := args[0]
	args = args[1:]

	if len(args) < requiredArgsSize {
		return conf, fmt.Errorf(
			"required %d arguments\n%s",
			requiredArgsSize,
			usage(cmd),
		)
	}

	conf.DbPath = args[argsDbPath]
	conf.Cmd = args[argsCmd]
	conf.Args = args[argsCmd+1:]

	err = validateDbPath(conf.DbPath)

	return
}
