package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"github.com/pomelo-studios/pomelo-hook/server/config"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func runInit() error {
	cfg := config.Load()
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	var count int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM organizations").Scan(&count); err != nil {
		return fmt.Errorf("check existing organizations: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("already initialized — use the admin panel to manage users")
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Organization name: ")
	orgName, err := scanLine(scanner, "organization name")
	if err != nil {
		return err
	}
	if orgName == "" {
		return fmt.Errorf("organization name required")
	}

	fmt.Print("Admin name: ")
	adminName, err := scanLine(scanner, "admin name")
	if err != nil {
		return err
	}
	if adminName == "" {
		return fmt.Errorf("admin name required")
	}

	fmt.Print("Admin email: ")
	adminEmail, err := scanLine(scanner, "admin email")
	if err != nil {
		return err
	}
	if adminEmail == "" {
		return fmt.Errorf("admin email required")
	}

	fmt.Print("Admin password (min 8 chars): ")
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read password: %w", err)
	}
	if len(passBytes) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword(passBytes, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	org, err := db.CreateOrg(orgName)
	if err != nil {
		return fmt.Errorf("create org: %w", err)
	}

	user, err := db.CreateUser(store.CreateUserParams{
		OrgID:        org.ID,
		Email:        adminEmail,
		Name:         adminName,
		Role:         "admin",
		PasswordHash: string(hash),
	})
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	fmt.Printf("\nOrganization '%s' created (id: %s)\n", org.Name, org.ID)
	fmt.Printf("Admin '%s' created.\n", user.Name)
	fmt.Printf("API Key: %s\n", user.APIKey)
	fmt.Println("Save this key — it won't be shown again. You can also run: pomelo-hook login")
	return nil
}

func scanLine(s *bufio.Scanner, field string) (string, error) {
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return "", fmt.Errorf("read %s: %w", field, err)
		}
		return "", fmt.Errorf("read %s: unexpected EOF", field)
	}
	return strings.TrimSpace(s.Text()), nil
}
