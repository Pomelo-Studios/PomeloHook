package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Tunnel struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	UserID       string `json:"user_id"`
	OrgID        string `json:"org_id"`
	Subdomain    string `json:"subdomain"`
	DisplayName  string `json:"display_name,omitempty"`
	ActiveUserID string `json:"active_user_id"`
	ActiveDevice string `json:"active_device"`
	Status       string `json:"status"`
}

type CreateTunnelParams struct {
	Type   string // "personal" | "org"
	UserID string // set for personal
	OrgID  string // set for org
	Name   string // optional, used as subdomain for org tunnels
}

const tunnelColumns = `id, type, COALESCE(user_id,''), COALESCE(org_id,''), subdomain, COALESCE(display_name,''), COALESCE(active_user_id,''), COALESCE(active_device,''), status`

func scanTunnel(row rowScanner) (*Tunnel, error) {
	t := &Tunnel{}
	return t, row.Scan(&t.ID, &t.Type, &t.UserID, &t.OrgID, &t.Subdomain, &t.DisplayName, &t.ActiveUserID, &t.ActiveDevice, &t.Status)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Store) CreateTunnel(p CreateTunnelParams) (*Tunnel, error) {
	id := uuid.NewString()
	subdomain := p.Name
	if subdomain == "" {
		var err error
		subdomain, err = randomHex(4)
		if err != nil {
			return nil, err
		}
	}
	_, err := s.db.Exec(
		`INSERT INTO tunnels (id, type, user_id, org_id, subdomain) VALUES (?,?,?,?,?)`,
		id, p.Type, nilIfEmpty(p.UserID), nilIfEmpty(p.OrgID), subdomain,
	)
	if err != nil {
		return nil, err
	}
	return &Tunnel{ID: id, Type: p.Type, UserID: p.UserID, OrgID: p.OrgID, Subdomain: subdomain, Status: "inactive"}, nil
}

func (s *Store) GetPersonalTunnel(userID string) (*Tunnel, error) {
	row := s.db.QueryRow(
		`SELECT `+tunnelColumns+` FROM tunnels WHERE user_id=? AND type='personal' LIMIT 1`, userID)
	t, err := scanTunnel(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (s *Store) GetTunnelBySubdomain(subdomain string) (*Tunnel, error) {
	row := s.db.QueryRow(`SELECT `+tunnelColumns+` FROM tunnels WHERE subdomain = ?`, subdomain)
	return scanTunnel(row)
}

func (s *Store) GetTunnelByID(id string) (*Tunnel, error) {
	row := s.db.QueryRow(`SELECT `+tunnelColumns+` FROM tunnels WHERE id = ?`, id)
	return scanTunnel(row)
}

func (s *Store) SetTunnelActive(tunnelID, userID, device string) error {
	_, err := s.db.Exec(
		`UPDATE tunnels SET active_user_id=?, active_device=?, status='active' WHERE id=?`,
		userID, nilIfEmpty(device), tunnelID,
	)
	return err
}

func (s *Store) SetTunnelInactive(tunnelID string) error {
	_, err := s.db.Exec(
		`UPDATE tunnels SET active_user_id=NULL, active_device=NULL, status='inactive' WHERE id=?`,
		tunnelID,
	)
	return err
}

func (s *Store) GetActiveTunnelUser(tunnelID string) (string, error) {
	var userID string
	err := s.db.QueryRow(`SELECT COALESCE(active_user_id,'') FROM tunnels WHERE id=?`, tunnelID).Scan(&userID)
	return userID, err
}

func (s *Store) ListTunnelsForUser(userID, orgID string) ([]*Tunnel, error) {
	rows, err := s.db.Query(
		`SELECT `+tunnelColumns+` FROM tunnels WHERE user_id=? OR org_id=?`,
		userID, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tunnels []*Tunnel
	for rows.Next() {
		t, err := scanTunnel(rows)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tunnels, nil
}

func (s *Store) ListOrgTunnels(orgID string) ([]*Tunnel, error) {
	rows, err := s.db.Query(
		`SELECT `+tunnelColumns+` FROM tunnels WHERE org_id=? AND type='org'`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tunnels []*Tunnel
	for rows.Next() {
		t, err := scanTunnel(rows)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	return tunnels, rows.Err()
}

func (s *Store) UpdateTunnelDisplayName(id, displayName string) (*Tunnel, error) {
	_, err := s.db.Exec(`UPDATE tunnels SET display_name = ? WHERE id = ?`, nilIfEmpty(displayName), id)
	if err != nil {
		return nil, err
	}
	return s.GetTunnelByID(id)
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

var ErrSubdomainTaken = errors.New("subdomain already taken by another user")

func (s *Store) GetOrCreatePersonalTunnel(userID, name string) (*Tunnel, bool, error) {
	// Always check for an existing personal tunnel first, regardless of name.
	// A user can only have one personal tunnel; name is only used when creating.
	existing, err := s.GetPersonalTunnel(userID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	if name != "" {
		// Ensure the desired subdomain is available before creating.
		_, err := s.GetTunnelBySubdomain(name)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, false, err
		}
		if err == nil {
			return nil, false, ErrSubdomainTaken
		}
	}

	tun, err := s.CreateTunnel(CreateTunnelParams{Type: "personal", UserID: userID, Name: name})
	if err != nil {
		// Handle race: another request created the tunnel between our GET and INSERT.
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			existing, err2 := s.GetPersonalTunnel(userID)
			if err2 == nil && existing != nil {
				return existing, false, nil
			}
			if err2 == nil {
				// Collision caused by another user's tunnel winning the race on this subdomain.
				return nil, false, ErrSubdomainTaken
			}
			return nil, false, err2
		}
		return nil, false, err
	}
	return tun, true, nil
}
