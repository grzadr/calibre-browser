package booksdb

import (
	"fmt"
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

type CommandFunc func(index *BookEntries, args []string) (BookEntrySlice, error)

func UnknownCommand(index *BookEntries, args []string) (BookEntrySlice, error) {
	return nil, fmt.Errorf("unknown command")
}

func SearchTitleCommand(
	index *BookEntries,
	args []string,
) (entries BookEntrySlice, err error) {
	found := index.titlesIndex.findSimilar(normalizeWordSlice(args))
	entries = make(BookEntrySlice, len(found))

	for i, bookId := range found {
		entries[i] = index.books[bookId]
	}

	return entries, nil
}

var CommandMap = [...]CommandFunc{
	UnknownCommand, SearchTitleCommand,
}
