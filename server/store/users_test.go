package store_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestCreateAndGetUser(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()

	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")

	user, err := db.CreateUser(store.CreateUserParams{
		OrgID: "org1",
		Email: "yagiz@example.com",
		Name:  "Yagiz",
		Role:  "admin",
	})
	require.NoError(t, err)
	require.NotEmpty(t, user.APIKey)
	require.True(t, strings.HasPrefix(user.APIKey, "ph_"))
	require.Len(t, user.APIKey, 51) // "ph_" (3) + 48 hex chars
	require.Equal(t, "yagiz@example.com", user.Email)

	found, err := db.GetUserByAPIKey(user.APIKey)
	require.NoError(t, err)
	require.Equal(t, user.ID, found.ID)
}
