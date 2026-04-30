package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"

	"github.com/google/uuid"
)

type User struct {
	ID           string
	OrgID        string
	Email        string
	Name         string
	APIKey       string
	Role         string
	PasswordHash string
}

type CreateUserParams struct {
	OrgID        string
	Email        string
	Name         string
	Role         string
	PasswordHash string
}

func (s *Store) CreateUser(p CreateUserParams) (*User, error) {
	id := uuid.NewString()
	key, err := generateAPIKey()
	if err != nil {
		return nil, err
	}
	_, err = s.DB.Exec(
		`INSERT INTO users (id, org_id, email, name, api_key, role, password_hash) VALUES (?,?,?,?,?,?,?)`,
		id, p.OrgID, p.Email, p.Name, key, p.Role, p.PasswordHash,
	)
	if err != nil {
		return nil, err
	}
	return &User{ID: id, OrgID: p.OrgID, Email: p.Email, Name: p.Name, APIKey: key, Role: p.Role, PasswordHash: p.PasswordHash}, nil
}

func (s *Store) GetUserByID(id, orgID string) (*User, error) {
	row := s.DB.QueryRow(
		`SELECT id, org_id, email, name, api_key, role, password_hash FROM users WHERE id=? AND org_id=?`,
		id, orgID,
	)
	u := &User{}
	return u, row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role, &u.PasswordHash)
}

func (s *Store) GetUserByAPIKey(key string) (*User, error) {
	row := s.DB.QueryRow(
		`SELECT id, org_id, email, name, api_key, role, password_hash FROM users WHERE api_key = ?`, key,
	)
	u := &User{}
	return u, row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role, &u.PasswordHash)
}

func (s *Store) GetUserByEmail(email string) (*User, error) {
	row := s.DB.QueryRow(
		`SELECT id, org_id, email, name, api_key, role, password_hash FROM users WHERE email = ?`, email,
	)
	u := &User{}
	return u, row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role, &u.PasswordHash)
}

func (s *Store) ListOrgUsers(orgID string) ([]*User, error) {
	rows, err := s.DB.Query(
		`SELECT id, org_id, email, name, api_key, role, password_hash FROM users WHERE org_id=?`, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role, &u.PasswordHash); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Store) SetPasswordHash(id, orgID, hash string) error {
	res, err := s.DB.Exec(`UPDATE users SET password_hash=? WHERE id=? AND org_id=?`, hash, id, orgID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func generateAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "ph_" + hex.EncodeToString(b), nil
}
