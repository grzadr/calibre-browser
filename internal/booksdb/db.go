package booksdb

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	"github.com/grzadr/calibre-browser/internal/model"
	_ "modernc.org/sqlite"
)

type (
	Word           string
	BookEntrySlice []model.BookEntryRow
	BookEntryId    int
)

const defaultIndexWordsCapacity = 1000 * 1024 // TODO Adjust for actual usage

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

type BookEntries struct {
	books  BookEntrySlice
	titles *BookSearchIndex
	// ids   map[types.BookId]BookEntryId
}

func NewTitleIndex(books BookEntrySlice) (index *BookSearchIndex) {
	index = NewBookSearchIndex(len(books))

	for id, book := range books {
		// id := book.ID
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

func NewBookEntries(
	repo *BookRepository,
	ctx context.Context,
) (*BookEntries, error) {
	entries := &BookEntries{}
	var err error

	if entries.books, err = repo.BookEntry(ctx); err != nil {
		return nil, fmt.Errorf("error listing books %q: %w", repo.dbPath, err)
	}

	entries.titles = NewTitleIndex(entries.books)

	return entries, nil
}

// type BooksDb struct {
// 	dbPath     string
// 	queries    *model.Queries
// 	books      map[int]*model.BookEntryRow
// 	titleIndex *TitleIndex
// }

// func NewBooksDb(dbPath string, ctx context.Context) (db *BooksDb, err error)
// {
// 	db = &BooksDb{dbPath: dbPath}
// 	var sqlDb *sql.DB
// 	sqlDb, err = sql.Open("sqlite", dbPath)
// 	if err != nil {
// 		err = fmt.Errorf("error opening db %q: %w", dbPath, err)
// 		return
// 	}

// 	db.queries = model.New(sqlDb)

// 	books, err := db.queries.BookEntry(ctx)
// 	if err != nil {
// 		err = fmt.Errorf("error listing books %q: %w", dbPath, err)
// 	}

// 	db.books = make(map[int]*model.BookEntryRow, len(books))

// 	for _, book := range books {
// 		db.books[int(book.ID)] = book
// 	}

// 	db.titleIndex = NewTitleIndex(db)

// 	return
// }

// type BookEntriesIndex struct {
// 	entries *BookEntries
// 	titles  *BookSearchIndex
// }

var (
	// db         *BooksDb.
	dbOnce     sync.Once
	repository *BookRepository
	index      atomic.Pointer[BookEntries]
	// entries    atomic.Pointer[BookEntries]
	// titleIndex atomic.Pointer[BookSearchIndex]
	// dbLock     sync.RWMutex.
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
