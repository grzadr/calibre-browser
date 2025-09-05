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
	counter := make(map[int]int)
	words := lowerCase(args)

	for _, word := range words {
		if ids, found := db.titleIndex.words[word]; found {
			for _, id := range ids {
				counter[id] += 1
			}
		}
	}

	type JaccardIndex struct {
		score float32
		id    int
	}

	scores := make([]JaccardIndex, len(counter))

	wordsSize := len(words)

	for id, count := range counter {
		scores = append(scores, JaccardIndex{
			score: count) / (wordsSize + db.titleIndex.sizes[id]),
		})
	}

	return BookEntrySlice{}, nil
}

var CommandMap = [...]CommandFunc{
	UnknownCommand, SearchTitleCommand,
}
