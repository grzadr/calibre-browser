-- name: ListAllBooks :many
SELECT
    id,
    title,
    author_sort AS authors,
    timestamp AS added_at,
    last_modified AS modified_at,
    path
FROM books;
