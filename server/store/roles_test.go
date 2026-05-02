package store_test

import (
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/stretchr/testify/require"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func TestGetRolePermissions(t *testing.T) {
	s := openTestStore(t)

	perms, err := s.GetRolePermissions("member")
	require.NoError(t, err)
	require.True(t, perms["view_events"], "member should have view_events")
	require.True(t, perms["replay_events"], "member should have replay_events")
	require.False(t, perms["create_org_tunnel"], "member should NOT have create_org_tunnel")

	// non-existent role returns empty map, no error
	empty, err := s.GetRolePermissions("nonexistent")
	require.NoError(t, err)
	require.Empty(t, empty)
}

func TestListRoles(t *testing.T) {
	s := openTestStore(t)
	roles, err := s.ListRoles()
	require.NoError(t, err)
	require.Len(t, roles, 4)
}

func TestCreateUpdateDeleteRole(t *testing.T) {
	s := openTestStore(t)

	role, err := s.CreateRole("viewer", "Viewer", []string{"view_events"})
	require.NoError(t, err)
	require.Equal(t, "viewer", role.Name)
	require.Equal(t, "Viewer", role.DisplayName)
	require.Equal(t, []string{"view_events"}, role.Permissions)
	require.False(t, role.IsSystem)

	updated, err := s.UpdateRole("viewer", "Read Only", []string{"view_events", "replay_events"})
	require.NoError(t, err)
	require.Equal(t, "Read Only", updated.DisplayName)
	require.Len(t, updated.Permissions, 2)

	err = s.DeleteRole("viewer")
	require.NoError(t, err)

	// system role cannot be deleted
	err = s.DeleteRole("member")
	require.ErrorIs(t, err, store.ErrSystemRole)
}
