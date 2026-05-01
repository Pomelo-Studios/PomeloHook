// server/store/orgs_test.go
package store_test

import (
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/stretchr/testify/require"
)

func TestGetOrg(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	require.NoError(t, err)

	org, err := db.GetOrg("org1")
	require.NoError(t, err)
	require.Equal(t, "org1", org.ID)
	require.Equal(t, "Acme", org.Name)
}

func TestUpdateOrg(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	require.NoError(t, err)

	org, err := db.UpdateOrg("org1", "Acme Corp")
	require.NoError(t, err)
	require.Equal(t, "Acme Corp", org.Name)
}

func TestCreateOrg(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	org, err := db.CreateOrg("Test Org")
	require.NoError(t, err)
	require.NotEmpty(t, org.ID)
	require.Equal(t, "Test Org", org.Name)

	fetched, err := db.GetOrg(org.ID)
	require.NoError(t, err)
	require.Equal(t, org.ID, fetched.ID)
}
