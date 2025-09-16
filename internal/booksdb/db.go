package booksdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"unicode"

	"github.com/grzadr/calibre-browser/internal/model"
	_ "modernc.org/sqlite"
)

type (
	Word           string
	BookEntrySlice []model.BookEntryRow
	BookEntryId    uint16
)

const (
	defaultIndexWordsCapacity = 1000 * 1024 // TODO Adjust for actual usage
	defaultJaccardCapacity    = 64
)

type BookRepository struct {
	dbPath string
	*model.Queries
}

func NewBookRepository(
	dbPath string,
	ctx context.Context,
) (*BookRepository, error) {
	sqlDb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening db %q: %w", dbPath, err)
	}

	return &BookRepository{dbPath: dbPath, Queries: model.New(sqlDb)}, nil
}

var diacriticalMap = map[rune]string{
	// Polish diacriticals (lowercase only)
	'ą': "a", 'ć': "c", 'ę': "e", 'ł': "l",
	'ń': "n", 'ó': "o", 'ś': "s", 'ź': "z", 'ż': "z",
}

func normalizeWord(word string) Word {
	var result strings.Builder

	result.Grow(len(word) * 2)

	for _, r := range word {
		lowered := unicode.ToLower(r)
		if mapped, exists := diacriticalMap[lowered]; exists {
			result.WriteString(mapped)
		} else {
			result.WriteRune(lowered)
		}
	}
	return Word(result.String())
}

func normalizeWordSlice(words []string) (lowered []Word) {
	lowered = make([]Word, len(words))
	for i, word := range words {
		lowered[i] = normalizeWord(word)
	}

	return
}

type SimilarityIndexSOA[Id comparable] struct {
	counts  []uint16
	totals  []uint16
	ids     []Id
	visited map[Id]int
}

type SimilarityIndexScore[Id any] struct {
	id    Id
	count uint16
	total uint16
}

type SimilarityIndexAOS[Id comparable] struct {
	scores  []SimilarityIndexScore[Id]
	visited map[Id]int
}

type BookEntries struct {
	books       BookEntrySlice
	titlesIndex *BookSearchIndex
}

func NewBookEntries(
	repo *BookRepository,
	ctx context.Context,
) (*BookEntries, error) {
	entries := &BookEntries{}
	var err error

	if entries.books, err = repo.BookEntry(ctx); err != nil {
		return nil, fmt.Errorf("error listing books %q: %w", repo.dbPath, err)
	}

	titles := make([]string, len(entries.books))

	for id, entry := range entries.books {
		titles[id] = entry.Title
	}

	entries.titlesIndex = NewTitleIndex(titles)

	return entries, nil
}

var (
	repository *BookRepository
	index      atomic.Pointer[BookEntries]
)

func RefreshBookEntries(repo *BookRepository, ctx context.Context) error {
	entries, err := NewBookEntries(repo, ctx)
	if err != nil {
		return fmt.Errorf(
			"failed to refresh book entries %q: %w",
			repo.dbPath,
			err,
		)
	}
	index.Store(entries)

	return nil
}

func PopulateBooksRepository(dbPath string, ctx context.Context) (err error) {
	repository, err = NewBookRepository(dbPath, ctx)
	if err != nil {
		return fmt.Errorf(
			"failed to populate books repository %q: %w",
			dbPath,
			err,
		)
	}

	err = RefreshBookEntries(repository, ctx)

	return
}

func ExecuteCommand(
	index *BookEntries,
	cmd string,
	args []string,
) (BookEntrySlice, error) {
	return CommandMap[NewCommand(cmd)](index, args)
}

func GetBooksEntries() *BookEntries {
	return index.Load()
}
