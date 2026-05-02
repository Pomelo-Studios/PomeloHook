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

func TestMigration5_RolesAndDisplayName(t *testing.T) {
	s, err := store.Open(":memory:")
	require.NoError(t, err)
	defer s.Close()

	var count int
	require.NoError(t, s.QueryRaw(&count, `SELECT COUNT(*) FROM roles`), "roles table missing")
	require.Equal(t, 4, count, "want 4 seeded roles")

	var col int
	require.NoError(t, s.QueryRaw(&col, `SELECT COUNT(*) FROM pragma_table_info('tunnels') WHERE name='display_name'`))
	require.Equal(t, 1, col, "tunnels.display_name column missing")

	var isSystem bool
	require.NoError(t, s.QueryRaw(&isSystem, `SELECT is_system FROM roles WHERE name='member'`))
	require.True(t, isSystem, "member role must be a system role")

	var memberPerms string
	require.NoError(t, s.QueryRaw(&memberPerms, `SELECT permissions FROM roles WHERE name='member'`))
	require.Contains(t, memberPerms, "view_events", "member must have view_events permission")
}
