package arguments

import (
	"fmt"
	"log"
	"os"
)

const (
	requiredClientArgsSize = 2
	requiredServerArgsSize = 1
	argsDbPath             = 0
	argsClientCmd          = 1
)

type Config struct {
	DbPath string
	Cmd    string
	Args   []string
}

func usageServer(cmd string) string {
	return fmt.Sprintf("%s <db filename>", cmd)
}

func usageCLient(cmd string) string {
	return fmt.Sprintf("%s <db filename/address> <cmd> ...", cmd)
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

	if len(args) < requiredServerArgsSize {
		return conf, fmt.Errorf(
			"required %d arguments\n%s",
			requiredServerArgsSize,
			usageServer(cmd),
		)
	}

	conf.DbPath = args[argsDbPath]
	conf.Args = args[requiredServerArgsSize+1:]

	err = validateDbPath(conf.DbPath)

	return conf, err
}

func ParseArgsClient(args []string) (conf Config, err error) {
	log.Printf("parsing client arguments: %+v", args)

	cmd := args[0]
	args = args[1:]

	if len(args) < requiredClientArgsSize {
		return conf, fmt.Errorf(
			"required %d arguments\n%s",
			requiredClientArgsSize,
			usageCLient(cmd),
		)
	}

	conf.DbPath = args[argsDbPath]
	conf.Cmd = args[argsClientCmd]
	conf.Args = args[argsClientCmd+1:]

	err = validateDbPath(conf.DbPath)

	return conf, err
}
