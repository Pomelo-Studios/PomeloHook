package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

func Open(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{DB: db}, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS organizations (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS users (
			id         TEXT PRIMARY KEY,
			org_id     TEXT REFERENCES organizations(id),
			email      TEXT UNIQUE NOT NULL,
			name       TEXT NOT NULL,
			api_key    TEXT UNIQUE NOT NULL,
			role       TEXT NOT NULL DEFAULT 'member',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS tunnels (
			id             TEXT PRIMARY KEY,
			type           TEXT NOT NULL,
			user_id        TEXT REFERENCES users(id),
			org_id         TEXT REFERENCES organizations(id),
			subdomain      TEXT UNIQUE NOT NULL,
			active_user_id TEXT REFERENCES users(id),
			status         TEXT NOT NULL DEFAULT 'inactive',
			created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS webhook_events (
			id              TEXT PRIMARY KEY,
			tunnel_id       TEXT REFERENCES tunnels(id),
			received_at     DATETIME NOT NULL,
			method          TEXT NOT NULL,
			path            TEXT NOT NULL,
			headers         TEXT NOT NULL,
			request_body    TEXT,
			response_status INTEGER,
			response_body   TEXT,
			response_ms     INTEGER,
			forwarded       BOOLEAN NOT NULL DEFAULT FALSE,
			replayed_at     DATETIME
		);
	`)
	return err
}
