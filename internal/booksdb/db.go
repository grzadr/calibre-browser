package booksdb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/grzadr/calibre-browser/internal/model"
)

type BooksDb struct {
	dbPath  string
	queries *model.Queries
	books   []*model.BooksListRow
}

var (
	db     *BooksDb
	dbOnce sync.Once
)

func InitializeBooksDb(dbPath string, ctx context.Context) (err error) {
	dbOnce.Do(func() {
		db = &BooksDb{dbPath: dbPath}
		var sqlDb *sql.DB
		sqlDb, err = sql.Open("sqlite", dbPath)
		if err != nil {
			err = fmt.Errorf("error opening db %q: %w", dbPath, err)
			return
		}

		db.queries = model.New(sqlDb)

		if db.books, err = db.queries.BooksList(ctx); err != nil {
			err = fmt.Errorf("error listing books %q: %w", dbPath, err)
		}
	})

	return
}

func GetBooksDb() *BooksDb {
	return db
}
