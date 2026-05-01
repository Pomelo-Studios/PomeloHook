package store

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
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

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// migration holds a versioned schema change.
type migration struct {
	version int
	// sql is executed directly when non-empty.
	sql string
	// fn is called inside a transaction when sql is empty.
	fn func(tx *sql.Tx) error
}

// addColumnIfNotExists is a helper for migrations that need ALTER TABLE ADD COLUMN
// on SQLite, which does not support the IF NOT EXISTS clause on ALTER TABLE.
func addColumnIfNotExists(tx *sql.Tx, table, column, definition string) error {
	var count int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?`, table, column,
	).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := tx.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
}

// migrations is the ordered list of schema changes applied once each.
var migrations = []migration{
	{version: 1, sql: `
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
    `},
	{version: 2, fn: func(tx *sql.Tx) error {
		return addColumnIfNotExists(tx, "users", "password_hash", "TEXT NOT NULL DEFAULT ''")
	}},
	{version: 3, fn: func(tx *sql.Tx) error {
		return addColumnIfNotExists(tx, "tunnels", "active_device", "TEXT")
	}},
	// Migration 4: timestamps are stored as UTC RFC3339 strings ending in 'Z'.
	// The CHECK constraints intentionally reject other valid RFC3339 forms (e.g. +00:00 offsets).
	{version: 4, sql: `
        CREATE TABLE IF NOT EXISTS webhook_events_new (
            id              TEXT PRIMARY KEY,
            tunnel_id       TEXT REFERENCES tunnels(id),
            received_at     TEXT NOT NULL CHECK(received_at GLOB '????-??-??T??:??:??Z'),
            method          TEXT NOT NULL,
            path            TEXT NOT NULL,
            headers         TEXT NOT NULL,
            request_body    TEXT,
            response_status INTEGER,
            response_body   TEXT,
            response_ms     INTEGER,
            forwarded       BOOLEAN NOT NULL DEFAULT FALSE,
            replayed_at     TEXT CHECK(replayed_at IS NULL OR replayed_at GLOB '????-??-??T??:??:??Z')
        );
        INSERT INTO webhook_events_new SELECT * FROM webhook_events;
        DROP TABLE webhook_events;
        ALTER TABLE webhook_events_new RENAME TO webhook_events;
        CREATE INDEX IF NOT EXISTS idx_events_tunnel_received
            ON webhook_events (tunnel_id, received_at);
    `},
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version    INTEGER PRIMARY KEY,
            applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        )
    `); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.version, err)
		}
		// INSERT OR IGNORE atomically claims the migration slot before any reads.
		// If another process already inserted the row the INSERT is a no-op (0 rows
		// affected) and we skip cleanly without a read-to-write lock upgrade, which
		// can cause SQLITE_BUSY in WAL mode under concurrent starts.
		res, err := tx.Exec(`INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)`, m.version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("claim migration %d: %w", m.version, err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			tx.Rollback()
			continue // already applied by this or another process
		}
		if m.fn != nil {
			if err := m.fn(tx); err != nil {
				tx.Rollback()
				return fmt.Errorf("run migration %d: %w", m.version, err)
			}
		} else {
			if _, err := tx.Exec(m.sql); err != nil {
				tx.Rollback()
				return fmt.Errorf("run migration %d: %w", m.version, err)
			}
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}
	return nil
}

// ExecRaw executes a raw SQL statement. For use in tests only.
func (s *Store) ExecRaw(query string, args ...any) error {
	_, err := s.db.Exec(query, args...)
	return err
}

// QueryRaw executes a raw SQL query and scans a single value. For use in tests only.
func (s *Store) QueryRaw(dest any, query string, args ...any) error {
	return s.db.QueryRow(query, args...).Scan(dest)
}
