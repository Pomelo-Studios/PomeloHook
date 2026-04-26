package forward_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pomelo-studios/pomelo-hook/cli/forward"
	"github.com/stretchr/testify/require"
)

func TestForwardDeliversToLocalServer(t *testing.T) {
	received := make(chan string, 1)
	local := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer local.Close()

	f := forward.New(local.URL)
	payload, _ := json.Marshal(map[string]any{
		"event_id": "evt-1",
		"method":   "POST",
		"path":     "/payment",
		"headers":  `{}`,
		"body":     `{"amount":100}`,
	})
	result, err := f.Forward(payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.Equal(t, "/payment", <-received)
}
