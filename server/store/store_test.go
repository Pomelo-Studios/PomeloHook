package store_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestStore_WALModeEnabled(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var mode string
	if err := s.DB.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("expected journal_mode=wal, got %q", mode)
	}
}

func TestOpenCreatesSchema(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Test Org')")
	require.NoError(t, err)
}
