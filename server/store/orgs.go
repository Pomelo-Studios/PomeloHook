package store

import (
	"database/sql"

	"github.com/google/uuid"
)

type OrgMember struct {
	ID                    string `json:"ID"`
	Name                  string `json:"Name"`
	Email                 string `json:"Email"`
	Role                  string `json:"Role"`
	ActiveTunnelSubdomain string `json:"ActiveTunnelSubdomain"`
}

type Org struct {
	ID        string
	Name      string
	CreatedAt string
}

func (s *Store) CreateOrg(name string) (*Org, error) {
	id := "org_" + uuid.NewString()
	_, err := s.db.Exec(`INSERT INTO organizations (id, name) VALUES (?, ?)`, id, name)
	if err != nil {
		return nil, err
	}
	return &Org{ID: id, Name: name}, nil
}

func (s *Store) GetOrg(orgID string) (*Org, error) {
	row := s.db.QueryRow(`SELECT id, name, created_at FROM organizations WHERE id = ?`, orgID)
	o := &Org{}
	return o, row.Scan(&o.ID, &o.Name, &o.CreatedAt)
}

// OrgCount returns the number of organizations in the store.
func (s *Store) OrgCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM organizations").Scan(&count)
	return count, err
}

func (s *Store) ListOrgUsersWithStatus(orgID string) ([]*OrgMember, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.name, u.email, u.role,
		       COALESCE(t.subdomain, '') AS active_subdomain
		FROM users u
		LEFT JOIN tunnels t ON t.active_user_id = u.id AND t.status = 'active'
		WHERE u.org_id = ?
		ORDER BY u.name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []*OrgMember
	for rows.Next() {
		m := &OrgMember{}
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Role, &m.ActiveTunnelSubdomain); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *Store) UpdateOrg(id, name string) (*Org, error) {
	res, err := s.db.Exec(`UPDATE organizations SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetOrg(id)
}
