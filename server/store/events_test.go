package store_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestSaveAndGetEvent(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	setupTunnel(t, db)

	event, err := db.SaveEvent(store.SaveEventParams{
		TunnelID:    "tunnel-1",
		Method:      "POST",
		Path:        "/webhook",
		Headers:     `{"Content-Type":"application/json"}`,
		RequestBody: `{"event":"payment.completed"}`,
	})
	require.NoError(t, err)
	require.NotEmpty(t, event.ID)

	got, err := db.GetEvent(event.ID)
	require.NoError(t, err)
	require.Equal(t, "POST", got.Method)
	require.False(t, got.Forwarded)
}

func TestListEventsByTunnel(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	setupTunnel(t, db)

	for i := 0; i < 3; i++ {
		db.SaveEvent(store.SaveEventParams{TunnelID: "tunnel-1", Method: "POST", Path: "/", Headers: "{}", RequestBody: ""})
	}

	events, err := db.ListEvents("tunnel-1", 10)
	require.NoError(t, err)
	require.Len(t, events, 3)
}

func TestDeleteOldEvents(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	setupTunnel(t, db)

	db.DB.Exec(`INSERT INTO webhook_events (id, tunnel_id, received_at, method, path, headers, forwarded) VALUES ('old-1','tunnel-1',?,?,?,?,?)`,
		time.Now().AddDate(0, 0, -31).Format(time.RFC3339), "POST", "/", "{}", false)

	deleted, err := db.DeleteEventsOlderThan(30)
	require.NoError(t, err)
	require.Equal(t, int64(1), deleted)
}

func setupTunnel(t *testing.T, db *store.Store) {
	t.Helper()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	db.DB.Exec("INSERT INTO users (id, org_id, email, name, api_key, role) VALUES ('user1','org1','a@b.com','A','key1','admin')")
	db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('tunnel-1','org','org1','stripe')")
}
