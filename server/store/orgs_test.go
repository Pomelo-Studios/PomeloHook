// server/store/orgs_test.go
package store_test

import (
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/stretchr/testify/require"
)

func TestGetOrg(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")

	org, err := db.GetOrg("org1")
	require.NoError(t, err)
	require.Equal(t, "org1", org.ID)
	require.Equal(t, "Acme", org.Name)
}

func TestUpdateOrg(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")

	org, err := db.UpdateOrg("org1", "Acme Corp")
	require.NoError(t, err)
	require.Equal(t, "Acme Corp", org.Name)
}
