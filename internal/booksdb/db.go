package booksdb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/grzadr/calibre-browser/internal/model"
	_ "modernc.org/sqlite"
)

type (
	BookId int
	Word   string
)

const defaultIndexWordsCapacity = 1000 * 1024 // TODO Adjust for actual usage

var diacriticalMap = map[rune]string{
	// Polish diacriticals (lowercase only)
	'ą': "a", 'ć': "c", 'ę': "e", 'ł': "l",
	'ń': "n", 'ó': "o", 'ś': "s", 'ź': "z", 'ż': "z",
}

func normalizeWord(word string) string {
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
	return result.String()
}

func normalizeWordSlice(words []string) (lowered []string) {
	lowered = make([]string, len(words))
	for i, word := range words {
		lowered[i] = normalizeWord(word)

		if word == "społeczna" {
			log.Println(word, lowered[i])
		}
	}

	return
}

func splitTitle(title string) (words []string) {
	words = slices.DeleteFunc(
		strings.Split(title, " "),
		func(word string) bool {
			return word == ""
		},
	)
	return normalizeWordSlice(words)
}

type TitleIndex struct {
	words map[string][]int
	sizes map[int]int
}

func NewTitleIndex(db *BooksDb) (index *TitleIndex) {
	index = &TitleIndex{
		words: make(map[string][]int, defaultIndexWordsCapacity),
		sizes: make(map[int]int, len(db.books)),
	}

	for id, book := range db.books {
		for _, word := range splitTitle(book.Title) {
			index.sizes[id] = len(word)

			if ids, found := index.words[word]; found {
				index.words[word] = append(ids, id)
			} else {
				index.words[word] = []int{id}
			}
		}
	}
	return
}

type BooksDb struct {
	dbPath     string
	queries    *model.Queries
	books      map[int]*model.BookEntryRow
	titleIndex *TitleIndex
}

func NewBooksDb(dbPath string, ctx context.Context) (db *BooksDb, err error) {
	db = &BooksDb{dbPath: dbPath}
	var sqlDb *sql.DB
	sqlDb, err = sql.Open("sqlite", dbPath)
	if err != nil {
		err = fmt.Errorf("error opening db %q: %w", dbPath, err)
		return
	}

	db.queries = model.New(sqlDb)

	books, err := db.queries.BookEntry(ctx)
	if err != nil {
		err = fmt.Errorf("error listing books %q: %w", dbPath, err)
	}

	db.books = make(map[int]*model.BookEntryRow, len(books))

	for _, book := range books {
		db.books[int(book.ID)] = book
	}

	db.titleIndex = NewTitleIndex(db)

	return
}

var (
	db     *BooksDb
	dbLock sync.RWMutex
)

func InitializeBooksDb(dbPath string, ctx context.Context) (err error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	db, err = NewBooksDb(dbPath, ctx)
	if err != nil {
		return
	}

	return
}

func ExecuteCommand(
	db *BooksDb,
	cmd string,
	args []string,
) (BookEntrySlice, error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	return CommandMap[NewCommand(cmd)](db, args)
}

func GetBooksDb() *BooksDb {
	return db
}
