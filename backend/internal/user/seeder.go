package user

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// SeedAdminUser ensures an admin account exists, returning a status message.
func SeedAdminUser(ctx context.Context, db *pgxpool.Pool) string {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@example.com"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "AdminPassword123!"
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	// 1. Check if an admin user already exists
	var exists bool
	checkSQL := `SELECT EXISTS(SELECT 1 FROM users WHERE role = $1 OR email = $2)`
	err := db.QueryRow(ctx, checkSQL, "admin", adminEmail).Scan(&exists)
	if err != nil {
		msg := fmt.Sprintf("[WARN] Could not verify admin status: %v", err)
		log.Println(msg)
		return msg
	}

	if exists {
		msg := "[INFO] Admin user already exists. Skipping seeding."
		log.Println(msg)
		return msg
	}

	// 2. Hash the admin password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		msg := fmt.Sprintf("[WARN] Password hashing failed: %v", err)
		log.Println(msg)
		return msg
	}

	// 3. Insert the new admin user
	insertSQL := `
		INSERT INTO users (username, email, password, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO NOTHING;
	`
	_, err = db.Exec(ctx, insertSQL, adminUsername, adminEmail, string(hashedPassword), "admin")
	if err != nil {
		msg := fmt.Sprintf("[WARN] Admin insertion failed: %v", err)
		log.Println(msg)
		return msg
	}

	msg := fmt.Sprintf("[INFO] Initial admin user created successfully: %s", adminEmail)
	log.Println(msg)
	return msg
}
