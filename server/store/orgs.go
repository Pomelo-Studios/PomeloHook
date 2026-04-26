package store

import "database/sql"

type Org struct {
	ID        string
	Name      string
	CreatedAt string
}

func (s *Store) GetOrg(orgID string) (*Org, error) {
	row := s.DB.QueryRow(`SELECT id, name, created_at FROM organizations WHERE id = ?`, orgID)
	o := &Org{}
	return o, row.Scan(&o.ID, &o.Name, &o.CreatedAt)
}

func (s *Store) UpdateOrg(id, name string) (*Org, error) {
	res, err := s.DB.Exec(`UPDATE organizations SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetOrg(id)
}
