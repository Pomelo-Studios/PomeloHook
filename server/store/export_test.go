// export_test.go exposes internal Store methods for testing only.
// This file is compiled exclusively during test builds (Go ignores _test.go
// files in production), so none of these symbols appear in the production API.
package store

func (s *Store) AppliedMigrationCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&n)
	return n, err
}
