package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS media_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		release_year INTEGER NOT NULL DEFAULT 0,
		duration_minutes INTEGER NOT NULL DEFAULT 0,
		total_episodes INTEGER NOT NULL DEFAULT 0,
		watched_episodes INTEGER NOT NULL DEFAULT 0,
		avg_episode_duration INTEGER NOT NULL DEFAULT 0,
		total_pages INTEGER NOT NULL DEFAULT 0,
		read_pages INTEGER NOT NULL DEFAULT 0,
		avg_page_read_time INTEGER NOT NULL DEFAULT 0,
		description TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS queue_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		media_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'planned',
		priority INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (media_id) REFERENCES media_items(id) ON DELETE CASCADE,
		UNIQUE(user_id, media_id)
	);
	`

	_, err := db.Exec(schema)
	return err
}
