package booksdb

import (
	"cmp"
	"hash"
	"hash/fnv"
	"slices"
	"strings"
)

const (
	defaultIndexWordsCapacity = 1000 * 1024 // TODO Adjust for actual usage
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
	// numWords map[BookEntryId]Count
	hashes map[uint64][]BookEntryId
}

func NewBookSearchIndex(capacity int) *BookSearchIndex {
	return &BookSearchIndex{
		words: make(map[Word][]BookEntryId, defaultIndexWordsCapacity),
		// numWords: make(map[BookEntryId]Count, capacity),
		numWords: make([]Count, capacity),
		hashes:   make(map[uint64][]BookEntryId, defaultIndexWordsCapacity),
	}
}

func hashWord(word Word, hasher hash.Hash64) uint64 {
	hasher.Reset()
	hasher.Write([]byte(word))
	return hasher.Sum64()
}

func NewTitleIndex(titles []string) (index *BookSearchIndex) {
	index = NewBookSearchIndex(len(titles))

	hasher := fnv.New64a()

	for id, title := range titles {
		entryId := BookEntryId(id)
		for _, word := range splitTitle(title) {
			index.numWords[entryId] = Count(len(word))
			hash := hashWord(word, hasher)

			if ids, found := index.words[word]; found {
				index.words[word] = append(ids, entryId)
				index.hashes[hash] = append(ids, entryId)
			} else {
				index.words[word] = []BookEntryId{entryId}
				index.hashes[hash] = append(ids, entryId)
			}
		}
	}

	return
}

func (index *BookSearchIndex) findSimilarV1(
	words []Word,
) []BookEntryId {
	countedWords := make([]int, 0, 64)
	totalWords := make([]float32, 0, 64)
	countedIds := make([]BookEntryId, 0, 64)
	indices := make([]int, 0, 64)
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

func (index *BookSearchIndex) size() int {
	return len(index.numWords)
}

type SimilarityIndexCount struct {
	id    BookEntryId
	count Count
	total Count
}

type SimilarityIndexScore struct {
	id    BookEntryId
	score float32
}

func (index *BookSearchIndex) findSimilar(
	words []Word,
) []BookEntryId {
	capacity := min(max(64, index.size()/4), 16384)

	counts := make(
		[]SimilarityIndexCount,
		0,
		capacity,
	)
	visited := make(map[BookEntryId]int, capacity)
	lastEntryId := 0

	for _, word := range words {
		for _, bookId := range index.words[word] {
			entryId, found := visited[bookId]

			if found {
				counts[entryId].count++
			} else {
				counts = append(counts, SimilarityIndexCount{
					id:    bookId,
					count: 1,
					total: index.numWords[bookId],
				})
				visited[bookId] = lastEntryId
				lastEntryId++
			}
		}
	}

	scores := make([]SimilarityIndexScore, len(counts))
	querySize := Count(len(words))

	for i, count := range counts {
		scores[i] = SimilarityIndexScore{
			id: count.id,
			score: float32(
				count.count,
			) / float32(
				count.total+querySize-count.count,
			),
		}
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
