package store

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

func Open(dsn string) (*Store, error) {
	// Enable foreign keys and WAL mode via DSN for file DBs
	// For :memory: we use a special URI
	if dsn == ":memory:" {
		dsn = "file::memory:?mode=memory&_pragma=foreign_keys(1)"
	} else {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn = dsn + sep + "_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Set max open connections to 1 for SQLite (single writer)
	db.SetMaxOpenConns(1)

	// Ping to verify connection is valid and catch bad DSNs early
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("open sqlite %q: %w", dsn, err)
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS organizations (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS users (
			id            TEXT PRIMARY KEY,
			org_id        TEXT REFERENCES organizations(id),
			email         TEXT UNIQUE NOT NULL,
			name          TEXT NOT NULL,
			api_key       TEXT UNIQUE NOT NULL,
			role          TEXT NOT NULL DEFAULT 'member',
			password_hash TEXT NOT NULL DEFAULT '',
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS tunnels (
			id             TEXT PRIMARY KEY,
			type           TEXT NOT NULL,
			user_id        TEXT REFERENCES users(id),
			org_id         TEXT REFERENCES organizations(id),
			subdomain      TEXT UNIQUE NOT NULL,
			active_user_id TEXT REFERENCES users(id),
			active_device  TEXT,
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
		CREATE INDEX IF NOT EXISTS idx_events_tunnel_received
			ON webhook_events (tunnel_id, received_at);
		CREATE INDEX IF NOT EXISTS idx_tunnels_user_id ON tunnels (user_id);
		CREATE INDEX IF NOT EXISTS idx_tunnels_org_id  ON tunnels (org_id);
		CREATE INDEX IF NOT EXISTS idx_tunnels_status   ON tunnels (status);
	`)
	if err != nil {
		return err
	}

	var colCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='password_hash'`).Scan(&colCount); err != nil {
		return fmt.Errorf("check password_hash column: %w", err)
	}
	if colCount == 0 {
		if _, err := tx.Exec(`ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("migrate password_hash column: %w", err)
		}
	}

	var deviceColCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tunnels') WHERE name='active_device'`).Scan(&deviceColCount); err != nil {
		return fmt.Errorf("check active_device column: %w", err)
	}
	if deviceColCount == 0 {
		if _, err := tx.Exec(`ALTER TABLE tunnels ADD COLUMN active_device TEXT`); err != nil {
			return fmt.Errorf("migrate active_device column: %w", err)
		}
	}

	return tx.Commit()
}
