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

func TestListOrgUsersWithStatus_ShowsActiveTunnel(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	org, err := db.CreateOrg("Acme")
	require.NoError(t, err)

	user, err := db.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@b.com", Name: "Alice", Role: "member"})
	require.NoError(t, err)

	tun, err := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})
	require.NoError(t, err)

	err = db.SetTunnelActive(tun.ID, user.ID, "laptop")
	require.NoError(t, err)

	members, err := db.ListOrgUsersWithStatus(org.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, tun.Subdomain, members[0].ActiveTunnelSubdomain)
}

func TestListOrgUsersWithStatus_NoDuplicates(t *testing.T) {
	s := openTestStore(t)
	org, err := s.CreateOrg("Test Org")
	require.NoError(t, err)
	user, err := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@test.com", Name: "Alice", Role: "member"})
	require.NoError(t, err)

	// Create two tunnels, both marked active for the same user
	t1, err := s.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})
	require.NoError(t, err)
	t2, err := s.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID, Name: "second"})
	require.NoError(t, err)
	require.NoError(t, s.SetTunnelActive(t1.ID, user.ID, "device1"))
	require.NoError(t, s.SetTunnelActive(t2.ID, user.ID, "device2"))

	members, err := s.ListOrgUsersWithStatus(org.ID)
	require.NoError(t, err)
	require.Len(t, members, 1, "each user should appear exactly once regardless of active tunnel count")
}
