package store_test

import (
	"path/filepath"
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
	if err := s.QueryRaw(&mode, "PRAGMA journal_mode"); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("expected journal_mode=wal, got %q", mode)
	}
}

func TestStore_TunnelIndexesExist(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	wantIndexes := []string{
		"idx_tunnels_user_id",
		"idx_tunnels_org_id",
		"idx_tunnels_status",
	}
	for _, idx := range wantIndexes {
		var name string
		err := s.QueryRaw(&name, `SELECT name FROM sqlite_master WHERE type='index' AND name=?`, idx)
		if err != nil {
			t.Errorf("index %q not found: %v", idx, err)
		}
	}
}

func TestOpenCreatesSchema(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Test Org')")
	require.NoError(t, err)
}

func TestOpen_DSNWithExistingQueryParams(t *testing.T) {
	dir := t.TempDir()
	dsn := "file:" + filepath.Join(dir, "test.db") + "?cache=shared"
	db, err := store.Open(dsn)
	if err != nil {
		t.Fatalf("open DSN with existing query params: %v", err)
	}
	defer db.Close()
}

func TestMigration_IsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "idempotent.db")

	db1, err := store.Open(path)
	require.NoError(t, err)
	db1.Close()

	// Open the same DB a second time — migration must not error
	db2, err := store.Open(path)
	require.NoError(t, err)
	db2.Close()
}

func TestMigration_VersionsAreRecorded(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	count, err := db.AppliedMigrationCount()
	require.NoError(t, err)
	require.Greater(t, count, 0, "at least one migration must have been recorded")
}
