package store_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestCreatePersonalTunnel(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

	tunnel, err := db.CreateTunnel(store.CreateTunnelParams{
		Type:   "personal",
		UserID: user.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "personal", tunnel.Type)
	require.NotEmpty(t, tunnel.Subdomain)
}

func TestOrgTunnelConflict(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

	tunnel, _ := db.CreateTunnel(store.CreateTunnelParams{
		Type:  "org",
		OrgID: "org1",
		Name:  "stripe",
	})

	err := db.SetTunnelActive(tunnel.ID, user.ID)
	require.NoError(t, err)

	active, err := db.GetActiveTunnelUser(tunnel.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, active)
}
