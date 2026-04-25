package store_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestOpenCreatesSchema(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Test Org')")
	require.NoError(t, err)
}
