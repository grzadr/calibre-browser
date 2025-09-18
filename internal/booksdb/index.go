package booksdb

import (
	"cmp"
	"slices"
	"strings"
)

const (
	defaultIndexWordsCapacity     = 1000 * 1024
	defaultMaxWordCounterCapacity = 16384
	defaultMinWordCounterCapacity = 64
	defaultCapacityDivisor        = 4
)

type Count uint8

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
	numWords []Count
}

func NewBookSearchIndex(capacity int) *BookSearchIndex {
	return &BookSearchIndex{
		words:    make(map[Word][]BookEntryId, defaultIndexWordsCapacity),
		numWords: make([]Count, capacity),
	}
}

func NewTitleIndex(titles []string) (index *BookSearchIndex) {
	index = NewBookSearchIndex(len(titles))

	for id, title := range titles {
		entryId := BookEntryId(id)

		for _, word := range splitTitle(title) {
			index.numWords[entryId] = Count(len(word))

			if ids, found := index.words[word]; found {
				index.words[word] = append(ids, entryId)
			} else {
				index.words[word] = []BookEntryId{entryId}
			}
		}
	}

	return
}

func (index *BookSearchIndex) size() int {
	return len(index.numWords)
}

type SimilarityIndexScore struct {
	id    BookEntryId
	score float32
}

func (index *BookSearchIndex) findSimilar(
	words []Word,
) []BookEntryId {
	capacity := min(
		max(defaultMinWordCounterCapacity, index.size()/defaultCapacityDivisor),
		defaultMaxWordCounterCapacity,
	)

	counts := make(map[BookEntryId]Count, capacity)

	for _, word := range words {
		for _, bookId := range index.words[word] {
			counts[bookId]++
		}
	}

	scores := make([]SimilarityIndexScore, len(counts))
	querySize := Count(len(words))
	i := 0

	for bookId, count := range counts {
		scores[i] = SimilarityIndexScore{
			id: bookId,
			score: float32(
				count,
			) / float32(
				index.numWords[bookId]+querySize-count,
			),
		}
		i++
	}

	slices.SortFunc(
		scores,
		func(left, right SimilarityIndexScore) int {
			return cmp.Compare(right.score, left.score)
		},
	)

	found := make([]BookEntryId, len(scores))

	for i, score := range scores {
		found[i] = score.id
	}

	return found
}
