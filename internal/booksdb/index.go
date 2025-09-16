package booksdb

import (
	"cmp"
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

type BookSearchIndex struct {
	words    map[Word][]BookEntryId
	numWords map[BookEntryId]int
}

func NewBookSearchIndex(capacity int) *BookSearchIndex {
	return &BookSearchIndex{
		words:    make(map[Word][]BookEntryId, defaultIndexWordsCapacity),
		numWords: make(map[BookEntryId]int, capacity),
	}
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

func (index *BookSearchIndex) findSimilar(
	words []Word,
) []BookEntryId {
	countedWords := make([]int, 0, defaultJaccardCapacity)
	totalWords := make([]float32, 0, defaultJaccardCapacity)
	countedIds := make([]BookEntryId, 0, defaultJaccardCapacity)
	indices := make([]int, 0, defaultJaccardCapacity)
	visitedIds := make(map[BookEntryId]int)

	lastIndex := 0

	for _, word := range words {
		ids, found := index.words[word]
		if !found {
			continue
		}
		for _, bookId := range ids {
			i, found := visitedIds[bookId]
			if !found {
				i = lastIndex
				lastIndex++
				countedWords = append(countedWords, 1)
				totalWords = append(totalWords, float32(index.numWords[bookId]))
				countedIds = append(countedIds, bookId)
				indices = append(indices, i)
				visitedIds[bookId] = i
			} else {
				countedWords[i]++
			}
		}
	}

	scores := make([]float32, len(countedIds))
	querySize := float32(len(words))

	for i := range countedIds {
		counted := float32(countedWords[i])

		scores[i] = counted / (querySize + totalWords[i] - counted)
	}

	slices.SortFunc(indices, func(left, right int) int {
		return cmp.Compare(scores[right], scores[left])
	})

	found := make([]BookEntryId, len(scores))

	for i, id := range indices {
		found[i] = countedIds[id]
	}

	return found
}
