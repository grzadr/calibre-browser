package booksdb

import (
	"cmp"
	"fmt"
	"slices"
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

func countWords(
	index *BookEntries,
	words []Word,
) (counted map[BookEntryId]int) {
	counted = make(map[BookEntryId]int)

	for _, word := range words {
		if ids, found := index.titles.words[word]; found {
			for _, id := range ids {
				counted[id] += 1
			}
		}
	}
	return
}

func SearchTitleCommand(
	index *BookEntries,
	args []string,
) (BookEntrySlice, error) {
	countedWords := countWords(index, normalizeWordSlice(args))

	if len(countedWords) == 0 {
		return BookEntrySlice{}, nil
	}

	numWords := len(args)

	type JaccardIndex struct {
		score float32
		id    BookEntryId
	}

	scores := make([]JaccardIndex, 0, len(countedWords))

	for id, count := range countedWords {
		scores = append(scores, JaccardIndex{
			score: float32(
				count,
			) / float32(
				numWords+index.titles.numWords[id]-count,
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
