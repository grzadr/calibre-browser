package booksdb

import (
	"math/rand"
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

		for j := 0; j < numWords; j++ {
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

func BenchmarkFindSimilar(b *testing.B) {
	const (
		numBooks       = 10000
		maxTitleLength = 12
		querySize      = 5
	)

	titles := generateTitles(numBooks, maxTitleLength)
	index := NewTitleIndex(titles)
	query := generateQueryWords(querySize)

	var result []BookEntryId

	for b.Loop() {
		result = index.findSimilar(query)
	}

	_ = result
}

func BenchmarkFindSimilarScaling(b *testing.B) {
	algorithms := []struct {
		name string
		fn   func(*BookSearchIndex, []Word) []BookEntryId
	}{
		{"V1", (*BookSearchIndex).findSimilar},
		{"V2", (*BookSearchIndex).findSimilarV2},
	}

	testCases := []struct {
		name           string
		numBooks       int
		maxTitleLength int
		querySize      int
	}{
		{"Small_100books_5query", 100, 8, 5},
		{"Medium_1000books_10query", 1000, 10, 10},
		{"Large_10000books_15query", 10000, 12, 15},
		{"XLarge_50000books_20query", 50000, 15, 20},
	}

	for _, alg := range algorithms {
		for _, tc := range testCases {
			b.Run(alg.name+"_"+tc.name, func(b *testing.B) {
				titles := generateTitles(tc.numBooks, tc.maxTitleLength)
				index := NewTitleIndex(titles)
				query := generateQueryWords(tc.querySize)

				var result []BookEntryId

				for b.Loop() {
					result = alg.fn(index, query)
				}

				_ = result
			})
		}
	}
}
