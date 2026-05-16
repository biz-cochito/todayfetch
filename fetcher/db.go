package fetcher

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func InitDB(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT,
		year INTEGER,
		text TEXT,
		url TEXT
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return db, nil
}
