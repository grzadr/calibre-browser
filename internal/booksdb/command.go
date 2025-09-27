package booksdb

import (
	"fmt"
	"log"
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

func SelectEntriesByTitleCommand(
	entries *BookEntries,
	args []string,
) (selected BookEntrySlice, err error) {
	log.Printf("performing title search for %+v", args)
	found := entries.titlesIndex.findSimilar(normalizeWordSlice(args))
	selected = make(BookEntrySlice, len(found))

	for i, bookId := range found {
		selected[i] = entries.books[bookId]
	}

	return selected, nil
}

var CommandMap = [...]CommandFunc{
	UnknownCommand, SelectEntriesByTitleCommand,
}
