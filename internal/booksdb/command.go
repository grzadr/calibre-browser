package booksdb

import (
	"cmp"
	"fmt"
	"slices"

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

func countWords(words []string) (counted map[int]int) {
	counted = make(map[int]int)
	words = normalizeWordSlice(words)

	for _, word := range words {
		if ids, found := db.titleIndex.words[word]; found {
			for _, id := range ids {
				counted[id] += 1
			}
		}
	}
	return
}

func SearchTitleCommand(db *BooksDb, args []string) (BookEntrySlice, error) {
	countedWords := countWords(args)

	if len(countedWords) == 0 {
		return BookEntrySlice{}, nil
	}

	numWords := len(args)

	type JaccardIndex struct {
		score float32
		id    int
	}

	scores := make([]JaccardIndex, 0, len(countedWords))

	for id, count := range countedWords {
		scores = append(scores, JaccardIndex{
			score: float32(
				count,
			) / float32(
				numWords+db.titleIndex.sizes[id]-count,
			),
			id: id,
		})
	}

	slices.SortFunc(scores, func(left, right JaccardIndex) int {
		return -1 * cmp.Compare(left.score, right.score)
	})

	slices.Reverse(scores)

	entries := make(BookEntrySlice, len(scores))

	for i, score := range scores {
		entries[i] = db.books[score.id]
	}

	return entries, nil
}

var CommandMap = [...]CommandFunc{
	UnknownCommand, SearchTitleCommand,
}
