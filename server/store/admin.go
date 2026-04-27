package store

import (
	"database/sql"
	"fmt"
	"strings"
)

type TableInfo struct {
	Name     string `json:"name"`
	RowCount int    `json:"row_count"`
}

type TableResult struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

type QueryResult struct {
	Columns  []string `json:"columns"`
	Rows     [][]any  `json:"rows"`
	Affected int64    `json:"affected"`
}

var allowedTables = map[string]bool{
	"organizations":  true,
	"users":          true,
	"tunnels":        true,
	"webhook_events": true,
}

func (s *Store) UpdateUser(id, orgID, email, name, role string) (*User, error) {
	res, err := s.DB.Exec(`UPDATE users SET email=?, name=?, role=? WHERE id=? AND org_id=?`, email, name, role, id, orgID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, sql.ErrNoRows
	}
	row := s.DB.QueryRow(`SELECT id, org_id, email, name, api_key, role FROM users WHERE id=?`, id)
	u := &User{}
	return u, row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role)
}

func (s *Store) DeleteUser(id, orgID string) error {
	// verify target belongs to org before deleting
	var count int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE id=? AND org_id=?`, id, orgID).Scan(&count); err != nil || count == 0 {
		return sql.ErrNoRows
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM webhook_events WHERE tunnel_id IN (SELECT id FROM tunnels WHERE user_id=?)`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM tunnels WHERE user_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM users WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) RotateAPIKey(id, orgID string) (string, error) {
	key, err := generateAPIKey()
	if err != nil {
		return "", err
	}
	res, err := s.DB.Exec(`UPDATE users SET api_key=? WHERE id=? AND org_id=?`, key, id, orgID)
	if err != nil {
		return "", err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return "", sql.ErrNoRows
	}
	return key, nil
}

func (s *Store) ListAllTunnels(orgID string) ([]*Tunnel, error) {
	rows, err := s.DB.Query(
		`SELECT `+tunnelColumns+` FROM tunnels WHERE org_id=? OR user_id IN (SELECT id FROM users WHERE org_id=?)`,
		orgID, orgID,
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

func (s *Store) DeleteTunnel(id, orgID string) error {
	// verify tunnel belongs to org before deleting
	var count int
	if err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM tunnels WHERE id=? AND (org_id=? OR user_id IN (SELECT id FROM users WHERE org_id=?))`,
		id, orgID, orgID,
	).Scan(&count); err != nil || count == 0 {
		return sql.ErrNoRows
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM webhook_events WHERE tunnel_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM tunnels WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) TunnelBelongsToOrg(id, orgID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM tunnels WHERE id=? AND (org_id=? OR user_id IN (SELECT id FROM users WHERE org_id=?))`,
		id, orgID, orgID,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) ListTables() ([]TableInfo, error) {
	rows, err := s.DB.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tables := make([]TableInfo, 0, len(names))
	for _, name := range names {
		if !allowedTables[name] {
			continue
		}
		var count int
		s.DB.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, name)).Scan(&count) //nolint:gosec — name from sqlite_master, filtered by allowedTables
		tables = append(tables, TableInfo{Name: name, RowCount: count})
	}
	return tables, nil
}

func (s *Store) GetTableRows(name string, limit, offset int) (*TableResult, error) {
	if !allowedTables[name] {
		return nil, fmt.Errorf("table %q not found", name)
	}
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	rows, err := s.DB.Query(fmt.Sprintf(`SELECT * FROM %s LIMIT ? OFFSET ?`, name), limit, offset) //nolint:gosec — name whitelisted
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	qr, err := scanQueryRows(rows)
	if err != nil {
		return nil, err
	}
	return &TableResult{Columns: qr.Columns, Rows: qr.Rows}, nil
}

func (s *Store) RunQuery(query string) (*QueryResult, error) {
	upper := strings.TrimSpace(strings.ToUpper(query))
	isRead := strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "EXPLAIN") || strings.HasPrefix(upper, "PRAGMA")
	if isRead {
		rows, err := s.DB.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanQueryRows(rows)
	}
	res, err := s.DB.Exec(query)
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	return &QueryResult{Affected: affected}, nil
}

func scanQueryRows(rows *sql.Rows) (*QueryResult, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := &QueryResult{Columns: cols}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		result.Rows = append(result.Rows, vals)
	}
	return result, rows.Err()
}
