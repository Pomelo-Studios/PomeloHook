package store

import (
	"database/sql"

	"github.com/google/uuid"
)

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
