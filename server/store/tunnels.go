package store

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

type Tunnel struct {
	ID           string
	Type         string
	UserID       string
	OrgID        string
	Subdomain    string
	ActiveUserID string
	Status       string
}

type CreateTunnelParams struct {
	Type   string // "personal" | "org"
	UserID string // set for personal
	OrgID  string // set for org
	Name   string // optional, used as subdomain for org tunnels
}

const tunnelColumns = `id, type, COALESCE(user_id,''), COALESCE(org_id,''), subdomain, COALESCE(active_user_id,''), status`

func scanTunnel(row rowScanner) (*Tunnel, error) {
	t := &Tunnel{}
	return t, row.Scan(&t.ID, &t.Type, &t.UserID, &t.OrgID, &t.Subdomain, &t.ActiveUserID, &t.Status)
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
	_, err := s.DB.Exec(
		`INSERT INTO tunnels (id, type, user_id, org_id, subdomain) VALUES (?,?,?,?,?)`,
		id, p.Type, nilIfEmpty(p.UserID), nilIfEmpty(p.OrgID), subdomain,
	)
	if err != nil {
		return nil, err
	}
	return &Tunnel{ID: id, Type: p.Type, UserID: p.UserID, OrgID: p.OrgID, Subdomain: subdomain, Status: "inactive"}, nil
}

func (s *Store) GetTunnelBySubdomain(subdomain string) (*Tunnel, error) {
	row := s.DB.QueryRow(`SELECT `+tunnelColumns+` FROM tunnels WHERE subdomain = ?`, subdomain)
	return scanTunnel(row)
}

func (s *Store) GetTunnelByID(id string) (*Tunnel, error) {
	row := s.DB.QueryRow(`SELECT `+tunnelColumns+` FROM tunnels WHERE id = ?`, id)
	return scanTunnel(row)
}

func (s *Store) SetTunnelActive(tunnelID, userID string) error {
	_, err := s.DB.Exec(`UPDATE tunnels SET active_user_id=?, status='active' WHERE id=?`, userID, tunnelID)
	return err
}

func (s *Store) SetTunnelInactive(tunnelID string) error {
	_, err := s.DB.Exec(`UPDATE tunnels SET active_user_id=NULL, status='inactive' WHERE id=?`, tunnelID)
	return err
}

func (s *Store) GetActiveTunnelUser(tunnelID string) (string, error) {
	var userID string
	err := s.DB.QueryRow(`SELECT COALESCE(active_user_id,'') FROM tunnels WHERE id=?`, tunnelID).Scan(&userID)
	return userID, err
}

func (s *Store) ListTunnelsForUser(userID, orgID string) ([]*Tunnel, error) {
	rows, err := s.DB.Query(
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

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
