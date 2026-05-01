package store_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestCreatePersonalTunnel(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
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
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

	tunnel, _ := db.CreateTunnel(store.CreateTunnelParams{
		Type:  "org",
		OrgID: "org1",
		Name:  "stripe",
	})

	err := db.SetTunnelActive(tunnel.ID, user.ID, "")
	require.NoError(t, err)

	active, err := db.GetActiveTunnelUser(tunnel.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, active)
}

func TestSetTunnelActiveStoresDevice(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	err := db.SetTunnelActive(tun.ID, user.ID, "MONSTER-2352")
	require.NoError(t, err)

	got, err := db.GetTunnelByID(tun.ID)
	require.NoError(t, err)
	require.Equal(t, "active", got.Status)
	require.Equal(t, "MONSTER-2352", got.ActiveDevice)
}

func TestSetTunnelInactiveClearsDevice(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	db.SetTunnelActive(tun.ID, user.ID, "MONSTER-2352")
	err := db.SetTunnelInactive(tun.ID)
	require.NoError(t, err)

	got, err := db.GetTunnelByID(tun.ID)
	require.NoError(t, err)
	require.Equal(t, "inactive", got.Status)
	require.Equal(t, "", got.ActiveDevice)
}

func TestGetPersonalTunnel_NilBeforeCreate(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "member"})

	got, err := db.GetPersonalTunnel(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestGetPersonalTunnel_ReturnsExisting(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "member"})

	created, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	got, err := db.GetPersonalTunnel(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != created.ID {
		t.Fatalf("expected tunnel %s, got %+v", created.ID, got)
	}
}

func TestListOrgTunnels(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

	db.CreateTunnel(store.CreateTunnelParams{Type: "org", OrgID: "org1", Name: "payment-wh"})
	db.CreateTunnel(store.CreateTunnelParams{Type: "org", OrgID: "org1", Name: "order-wh"})
	db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID}) // must not appear

	tunnels, err := db.ListOrgTunnels("org1")
	require.NoError(t, err)
	require.Len(t, tunnels, 2)
	for _, tunnel := range tunnels {
		require.Equal(t, "org", tunnel.Type)
	}
}
