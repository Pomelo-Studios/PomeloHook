package store_test

import (
	"database/sql"
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

func TestSetPasswordHash(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('o1', 'T')")
	require.NoError(t, err)

	u, err := db.CreateUser(store.CreateUserParams{OrgID: "o1", Email: "x@y.com", Name: "X", Role: "admin"})
	require.NoError(t, err)

	err = db.SetPasswordHash(u.ID, "o1", "$hash$")
	require.NoError(t, err)

	fetched, err := db.GetUserByEmail("x@y.com")
	require.NoError(t, err)
	require.Equal(t, "$hash$", fetched.PasswordHash)
}

func TestSetPasswordHashWrongOrg(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('o1', 'T')")
	require.NoError(t, err)

	u, err := db.CreateUser(store.CreateUserParams{OrgID: "o1", Email: "x@y.com", Name: "X", Role: "admin"})
	require.NoError(t, err)

	err = db.SetPasswordHash(u.ID, "wrong-org", "$hash$")
	require.ErrorIs(t, err, sql.ErrNoRows)
}
