package booksdb

import (
	"fmt"

	"github.com/grzadr/calibre-browser/internal/model"
)

type Command byte

const (
	Unknown Command = iota
	SearchTitle
)

func NewCommand(cmd string) Command {
	switch cmd {
	case "title":
		return SearchTitle
	default:
		return Unknown
	}
}

type BookEntrySlice []*model.BookEntryRow

type CommandFunc func(db *BooksDb, args []string) (BookEntrySlice, error)

func UnknownCommand(db *BooksDb, args []string) (BookEntrySlice, error) {
	return nil, fmt.Errorf("unknown command")
}

func SearchTitleCommand(db *BooksDb, args []string) (BookEntrySlice, error) {
	return nil, nil
}

var CommandMap = [...]CommandFunc{
	UnknownCommand, SearchTitleCommand,
}
