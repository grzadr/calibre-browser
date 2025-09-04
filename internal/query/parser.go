package query

type Command byte

const (
	UnknownCommand Command = iota
	SearchTitle
)

func NewCommand(cmd string) Command {
	switch cmd {
	case "title":
		return SearchTitle
	default:
		return UnknownCommand
	}
}

type SearchQueryArgs []string
