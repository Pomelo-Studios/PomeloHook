package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WebhookEvent struct {
	ID             string
	TunnelID       string
	ReceivedAt     time.Time
	Method         string
	Path           string
	Headers        string
	RequestBody    string
	ResponseStatus int
	ResponseBody   string
	ResponseMS     int64
	Forwarded      bool
	ReplayedAt     *time.Time
}

type SaveEventParams struct {
	TunnelID    string
	Method      string
	Path        string
	Headers     string
	RequestBody string
}

const eventColumns = `id, tunnel_id, received_at, method, path, headers,
         COALESCE(request_body,''), COALESCE(response_status,0),
         COALESCE(response_body,''), COALESCE(response_ms,0), forwarded, replayed_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEvent(row rowScanner) (*WebhookEvent, error) {
	e := &WebhookEvent{}
	var receivedAt string
	var replayedAt *string
	if err := row.Scan(&e.ID, &e.TunnelID, &receivedAt, &e.Method, &e.Path, &e.Headers,
		&e.RequestBody, &e.ResponseStatus, &e.ResponseBody, &e.ResponseMS, &e.Forwarded, &replayedAt); err != nil {
		return nil, err
	}
	var parseErr error
	e.ReceivedAt, parseErr = time.Parse(time.RFC3339, receivedAt)
	if parseErr != nil {
		return nil, fmt.Errorf("parse received_at for event %s: %w", e.ID, parseErr)
	}
	if replayedAt != nil {
		if t, err := time.Parse(time.RFC3339, *replayedAt); err == nil {
			e.ReplayedAt = &t
		}
	}
	return e, nil
}

func (s *Store) SaveEvent(p SaveEventParams) (*WebhookEvent, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	_, err := s.DB.Exec(
		`INSERT INTO webhook_events (id, tunnel_id, received_at, method, path, headers, request_body) VALUES (?,?,?,?,?,?,?)`,
		id, p.TunnelID, now.Format(time.RFC3339), p.Method, p.Path, p.Headers, p.RequestBody,
	)
	if err != nil {
		return nil, err
	}
	return &WebhookEvent{
		ID: id, TunnelID: p.TunnelID, ReceivedAt: now,
		Method: p.Method, Path: p.Path, Headers: p.Headers, RequestBody: p.RequestBody,
	}, nil
}

func (s *Store) GetEvent(id string) (*WebhookEvent, error) {
	row := s.DB.QueryRow(`SELECT `+eventColumns+` FROM webhook_events WHERE id=?`, id)
	e, err := scanEvent(row)
	if err == sql.ErrNoRows {
		return nil, err
	}
	return e, err
}

func (s *Store) ListEvents(tunnelID string, limit int) ([]*WebhookEvent, error) {
	rows, err := s.DB.Query(
		`SELECT `+eventColumns+` FROM webhook_events WHERE tunnel_id=? ORDER BY received_at DESC LIMIT ?`,
		tunnelID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []*WebhookEvent
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Store) MarkEventForwarded(id string, status int, body string, ms int64) error {
	_, err := s.DB.Exec(
		`UPDATE webhook_events SET forwarded=TRUE, response_status=?, response_body=?, response_ms=? WHERE id=?`,
		status, body, ms, id,
	)
	return err
}

func (s *Store) MarkEventReplayed(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.Exec(`UPDATE webhook_events SET replayed_at=? WHERE id=?`, now, id)
	return err
}

func (s *Store) DeleteEventsOlderThan(days int) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)
	res, err := s.DB.Exec(`DELETE FROM webhook_events WHERE received_at < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
