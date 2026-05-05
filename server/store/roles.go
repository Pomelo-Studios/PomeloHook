package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

var ErrSystemRole = errors.New("system role cannot be deleted")

type Role struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetRolePermissions returns the permission map for a role visible to the given org
// (system roles are always visible; custom roles are scoped to the org).
func (s *Store) GetRolePermissions(roleName, orgID string) (map[string]bool, error) {
	var permJSON string
	err := s.db.QueryRow(
		`SELECT permissions FROM roles WHERE name = ? AND (is_system = TRUE OR org_id = ?)`,
		roleName, orgID,
	).Scan(&permJSON)
	if err == sql.ErrNoRows {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}
	var perms []string
	if err := json.Unmarshal([]byte(permJSON), &perms); err != nil {
		return map[string]bool{}, nil
	}
	out := make(map[string]bool, len(perms))
	for _, p := range perms {
		out[p] = true
	}
	return out, nil
}

func (s *Store) ListRoles(orgID string) ([]*Role, error) {
	rows, err := s.db.Query(
		`SELECT name, display_name, permissions, is_system, created_at FROM roles
		 WHERE is_system = TRUE OR org_id = ?
		 ORDER BY created_at`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*Role
	for rows.Next() {
		r, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

func (s *Store) GetRole(name, orgID string) (*Role, error) {
	row := s.db.QueryRow(
		`SELECT name, display_name, permissions, is_system, created_at FROM roles
		 WHERE name = ? AND (is_system = TRUE OR org_id = ?)`,
		name, orgID,
	)
	return scanRole(row)
}

func (s *Store) CreateRole(orgID, name, displayName string, permissions []string) (*Role, error) {
	if permissions == nil {
		permissions = []string{}
	}
	permJSON, _ := json.Marshal(permissions)
	_, err := s.db.Exec(
		`INSERT INTO roles (name, display_name, permissions, org_id) VALUES (?, ?, ?, ?)`,
		name, displayName, string(permJSON), orgID,
	)
	if err != nil {
		return nil, err
	}
	return s.GetRole(name, orgID)
}

func (s *Store) UpdateRole(orgID, name, displayName string, permissions []string) (*Role, error) {
	if permissions == nil {
		permissions = []string{}
	}
	permJSON, _ := json.Marshal(permissions)
	res, err := s.db.Exec(
		`UPDATE roles SET display_name = ?, permissions = ? WHERE name = ? AND org_id = ?`,
		displayName, string(permJSON), name, orgID,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetRole(name, orgID)
}

// DeleteRole removes a custom org role and falls back all members with that role to "member".
func (s *Store) DeleteRole(orgID, name string) error {
	var isSystem bool
	err := s.db.QueryRow(`SELECT is_system FROM roles WHERE name = ?`, name).Scan(&isSystem)
	if err == sql.ErrNoRows {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}
	if isSystem {
		return ErrSystemRole
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE users SET role = 'member' WHERE role = ? AND org_id = ?`, name, orgID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM roles WHERE name = ? AND org_id = ?`, name, orgID); err != nil {
		return err
	}
	return tx.Commit()
}

func scanRole(row rowScanner) (*Role, error) {
	r := &Role{}
	var permJSON string
	if err := row.Scan(&r.Name, &r.DisplayName, &permJSON, &r.IsSystem, &r.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(permJSON), &r.Permissions); err != nil {
		r.Permissions = []string{}
	}
	return r, nil
}
