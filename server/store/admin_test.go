// server/store/admin_test.go
package store_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/stretchr/testify/require"
)

func openWithOrg(t *testing.T) (*store.Store, *store.User) {
	t.Helper()
	db, _ := store.Open(":memory:")
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	u, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "Alice", Role: "admin"})
	return db, u
}

func TestDeleteUser_NotFound(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	_, err := db.DeleteUser("nonexistent-id", "org1")
	if err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}

func TestDeleteTunnel_NotFound(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	err := db.DeleteTunnel("nonexistent-id", "org1")
	if err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}

func TestDeleteUser_WrongOrg(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org2', 'Beta')")

	_, err := db.DeleteUser(u.ID, "org2")
	if err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows for wrong org, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()

	updated, err := db.UpdateUser(u.ID, "org1", "new@b.com", "Alice New", "member")
	require.NoError(t, err)
	require.Equal(t, "new@b.com", updated.Email)
	require.Equal(t, "Alice New", updated.Name)
	require.Equal(t, "member", updated.Role)
}

func TestDeleteUser(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()

	_, err := db.DeleteUser(u.ID, "org1")
	require.NoError(t, err)
	_, err = db.GetUserByEmail("a@b.com")
	require.Error(t, err)
}

func TestDeleteUser_ReturnsKeyFromTransaction(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()

	key, err := db.DeleteUser(u.ID, "org1")
	require.NoError(t, err)
	require.Equal(t, u.APIKey, key, "returned key must match the key that was active at deletion time")

	_, err = db.GetUserByEmail("a@b.com")
	require.Error(t, err)
}

func TestRotateAPIKey_NotFound(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	_, _, err := db.RotateAPIKey("nonexistent-id", "org1")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestRotateAPIKey_WrongOrg(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org2', 'Beta')")

	_, _, err := db.RotateAPIKey(u.ID, "org2")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for wrong org, got %v", err)
	}
}

func TestRotateAPIKey(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()

	oldKey, newKey, err := db.RotateAPIKey(u.ID, "org1")
	require.NoError(t, err)
	require.Equal(t, u.APIKey, oldKey)
	require.NotEqual(t, u.APIKey, newKey)
	require.Contains(t, newKey, "ph_")
}

func TestListAllTunnels(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()
	db.DB.Exec("INSERT INTO tunnels (id, type, user_id, subdomain) VALUES ('t1','personal',?,?)", u.ID, "abc")
	db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t2','org','org1','def')")

	tunnels, err := db.ListAllTunnels("org1")
	require.NoError(t, err)
	require.Len(t, tunnels, 2)
}

func TestDeleteTunnel(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()
	db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','abc')")

	require.NoError(t, db.DeleteTunnel("t1", "org1"))
	tunnels, _ := db.ListAllTunnels("org1")
	require.Empty(t, tunnels)
}

func TestListTables_ReturnsCounts(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	tables, err := db.ListTables()
	if err != nil {
		t.Fatal(err)
	}

	counts := map[string]int{}
	for _, tbl := range tables {
		counts[tbl.Name] = tbl.RowCount
	}

	if counts["organizations"] < 1 {
		t.Errorf("expected at least 1 org, got %d", counts["organizations"])
	}
	if counts["users"] < 1 {
		t.Errorf("expected at least 1 user, got %d", counts["users"])
	}
}

func TestListTables(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	tables, err := db.ListTables()
	require.NoError(t, err)
	names := make([]string, len(tables))
	for i, tbl := range tables {
		names[i] = tbl.Name
	}
	require.Contains(t, names, "users")
	require.Contains(t, names, "organizations")
}

func TestGetTableRows(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	result, err := db.GetTableRows("organizations", 10, 0)
	require.NoError(t, err)
	require.Contains(t, result.Columns, "id")
	require.Len(t, result.Rows, 1)
}

func TestGetTableRowsRejectsUnknownTable(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	_, err := db.GetTableRows("secret_table", 10, 0)
	require.Error(t, err)
}

func TestRunQuerySelect(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	result, err := db.RunQuery("SELECT id, name FROM organizations")
	require.NoError(t, err)
	require.Equal(t, []string{"id", "name"}, result.Columns)
	require.Len(t, result.Rows, 1)
}

func TestRunQueryWriteRejected(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	_, err := db.RunQuery("INSERT INTO organizations (id, name) VALUES ('org2', 'Beta')")
	require.Error(t, err)
	require.Contains(t, err.Error(), "only SELECT")
}

func TestRunQueryPragmaTableInfo(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	result, err := db.RunQuery("PRAGMA table_info(users)")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Columns)
}

func TestRunQueryPragmaPageCount(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	result, err := db.RunQuery("PRAGMA page_count")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestRunQueryPragmaWriteRejected(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	_, err := db.RunQuery("PRAGMA journal_mode=WAL")
	require.Error(t, err)
	require.Contains(t, err.Error(), "whitelisted read-only")
}

func TestRunQueryPragmaWriteWithParensRejected(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	_, err := db.RunQuery("PRAGMA journal_mode(WAL)")
	require.Error(t, err)
	require.Contains(t, err.Error(), "whitelisted read-only")
}

func TestRunQueryWithCTE(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	result, err := db.RunQuery("WITH t AS (SELECT id FROM organizations) SELECT * FROM t")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, []string{"id"}, result.Columns)
	require.Len(t, result.Rows, 1)
}
