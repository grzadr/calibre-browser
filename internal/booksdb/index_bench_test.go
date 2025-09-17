package booksdb

import (
	"cmp"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

// Common English words of varying lengths (1-12 letters) for test data
// generation
var testWords = [...]string{
	"a", "I", "be", "to", "of", "and", "in", "it", "you", "that", "he", "was",
	"for", "on", "are", "as", "with", "his", "they", "at", "have", "this",
	"from", "they", "she", "or", "one", "had", "by", "word", "but", "not",
	"what", "all", "were", "when", "your", "can", "said", "there", "each",
	"which", "their", "time", "will", "about", "if", "up", "out", "many",
	"then", "them", "these", "so", "some", "her", "would", "make", "like",
	"into", "him", "has", "two", "more", "very", "after", "words", "long",
	"just", "where", "through", "much", "before", "move", "right", "boy",
	"old", "too", "same", "tell", "does", "set", "three", "want", "air",
	"well", "also", "play", "small", "end", "put", "home", "read", "hand",
	"port", "large", "spell", "add", "even", "land",
}

const lenTestWords = len(testWords)

func generateTitles(numTitles, maxTitleLength int) []string {
	result := make([]string, numTitles)

	for i := range numTitles {
		numWords := rand.Intn(maxTitleLength) + 1
		words := make([]string, numWords)

		for j := range numWords {
			words[j] = testWords[rand.Intn(lenTestWords)]
		}

		result[i] = strings.Join(words, " ")
	}

	return result
}

func generateQueryWords(numWords int) []Word {
	words := make([]Word, numWords)

	for i := range numWords {
		words[i] = Word(testWords[rand.Intn(lenTestWords)])
	}

	return words
}

var testCases = []struct {
	name           string
	numBooks       int
	maxTitleLength int
	querySize      int
}{
	// {"Small_100books", 100, 12, 6},
	// {"Medium_1000books", 1000, 12, 6},
	{"Large_10000books", 10000, 12, 6},
	{"XLarge_50000books", 50000, 12, 6},
	// {"Large_10000books_3query", 10000, 12, 3},
	// {"Large_10000books_6query", 10000, 12, 6},
	// {"Large_10000books_9query", 10000, 12, 9},
	{"Large_10000books_12query", 10000, 12, 12},
	// {"Large_10000books_3title", 10000, 3, 6},
	// {"Large_10000books_6title", 10000, 6, 6},
	// {"Large_10000books_9title", 10000, 9, 6},
	// {"Large_10000books_12title", 10000, 12, 6},
}

// func BenchmarkFindSimilarScaling(b *testing.B) {
// 	algorithms := []struct {
// 		name string
// 		fn   func(*BookSearchIndex, []Word) []BookEntryId
// 	}{
// 		{"V1", (*BookSearchIndex).findSimilarV1},
// 		{"Base", (*BookSearchIndex).findSimilar},
// 		{"V3", (*BookSearchIndex).findSimilarV3},
// 		{"V4", (*BookSearchIndex).findSimilarV4},
// 	}

// 	for _, tc := range testCases {
// 		for _, alg := range algorithms {
// 			b.Run(alg.name+"_"+tc.name, func(b *testing.B) {
// 				titles := generateTitles(tc.numBooks, tc.maxTitleLength)
// 				index := NewTitleIndex(titles)
// 				query := generateQueryWords(tc.querySize)

// 				var result []BookEntryId

// 				for b.Loop() {
// 					result = alg.fn(index, query)
// 				}

// 				_ = result
// 			})
// 		}
// 	}
// }

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

func BenchmarkFindSimilarScaling(b *testing.B) {
	fn := (*BookSearchIndex).findSimilar

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			titles := generateTitles(tc.numBooks, tc.maxTitleLength)
			index := NewTitleIndex(titles)
			query := generateQueryWords(tc.querySize)

			var result []BookEntryId

			for b.Loop() {
				result = fn(index, query)
			}

			_ = result
		})
	}
}
