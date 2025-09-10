package booksdb

import (
	"slices"
	"strings"
)

func splitTitle(title string) []Word {
	return normalizeWordSlice(slices.DeleteFunc(
		strings.Split(title, " "),
		func(word string) bool {
			return word == ""
		},
	))
}

func NewTitleIndex(books BookEntrySlice) (index *BookSearchIndex) {
	index = NewBookSearchIndex(len(books))

	for id, book := range books {
		entryId := BookEntryId(id)
		for _, word := range splitTitle(book.Title) {
			index.numWords[entryId] = len(word)

			if ids, found := index.words[word]; found {
				index.words[word] = append(ids, entryId)
			} else {
				index.words[word] = []BookEntryId{entryId}
			}
		}
	}

	return
}
