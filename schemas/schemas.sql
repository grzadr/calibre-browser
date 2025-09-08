CREATE TABLE books (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT NOT NULL DEFAULT 'Unknown' COLLATE NOCASE,
    sort            TEXT COLLATE NOCASE,
    timestamp       TIMESTAMP NOT NULL,
    pubdate         TIMESTAMP NOT NULL,
    series_index    REAL NOT NULL DEFAULT 1.0,
    author_sort     TEXT NOT NULL COLLATE NOCASE,
    isbn            TEXT NOT NULL COLLATE NOCASE,
    lccn            TEXT NOT NULL COLLATE NOCASE,
    path            TEXT NOT NULL COLLATE NOCASE,
    flags           INTEGER NOT NULL DEFAULT 1,
    uuid            TEXT,
    has_cover       BOOL DEFAULT 0,
    last_modified   TIMESTAMP NOT NULL
);

