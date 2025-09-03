package arguments

import (
	"fmt"
	"os"
)

const (
	requiredArgsSize = 1
	argsDbPath       = 0
)

type Config struct {
	DbPath string
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

	conf.DbPath = args[argsDbPath]

	err = validateDbPath(conf.DbPath)

	return
}
