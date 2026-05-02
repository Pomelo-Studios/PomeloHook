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

func (s *Store) GetRolePermissions(roleName string) (map[string]bool, error) {
	var permJSON string
	err := s.db.QueryRow(`SELECT permissions FROM roles WHERE name = ?`, roleName).Scan(&permJSON)
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

func (s *Store) ListRoles() ([]*Role, error) {
	rows, err := s.db.Query(`SELECT name, display_name, permissions, is_system, created_at FROM roles ORDER BY created_at`)
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

func (s *Store) GetRole(name string) (*Role, error) {
	row := s.db.QueryRow(`SELECT name, display_name, permissions, is_system, created_at FROM roles WHERE name = ?`, name)
	return scanRole(row)
}

func (s *Store) CreateRole(name, displayName string, permissions []string) (*Role, error) {
	if permissions == nil {
		permissions = []string{}
	}
	permJSON, _ := json.Marshal(permissions)
	_, err := s.db.Exec(
		`INSERT INTO roles (name, display_name, permissions) VALUES (?, ?, ?)`,
		name, displayName, string(permJSON),
	)
	if err != nil {
		return nil, err
	}
	return s.GetRole(name)
}

func (s *Store) UpdateRole(name, displayName string, permissions []string) (*Role, error) {
	if permissions == nil {
		permissions = []string{}
	}
	permJSON, _ := json.Marshal(permissions)
	res, err := s.db.Exec(
		`UPDATE roles SET display_name = ?, permissions = ? WHERE name = ?`,
		displayName, string(permJSON), name,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetRole(name)
}

func (s *Store) DeleteRole(name string) error {
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
	_, err = s.db.Exec(`DELETE FROM roles WHERE name = ?`, name)
	return err
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
