package user

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleUser    = "user"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func InitUserTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin', 'manager', 'user')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	return err
}

func RegisterUser(db *sql.DB, username, email, rawPassword, role string) (*User, error) {
	if role == "" {
		role = RoleUser
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password failed: %w", err)
	}
	createdAt := time.Now().UTC()
	query := `INSERT INTO users (username, email, password, role, created_at) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query, username, email, string(hashedBytes), role, createdAt)
	if err != nil {
		return nil, fmt.Errorf("user insertion failed: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &User{
		ID:        id,
		Username:  username,
		Email:     email,
		Role:      role,
		CreatedAt: createdAt,
	}, nil
}

func GetUserByID(db *sql.DB, id int64) (*User, error) {
	query := `SELECT id, username, email, role, created_at FROM users WHERE id = ?`

	var u User
	err := db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return &u, nil
}
func UpdateUser(db *sql.DB, id int64, username, email, role string) (*User, error) {
	query := `UPDATE users SET username = ?, email = ?, role = ? WHERE id = ?`

	res, err := db.Exec(query, username, email, role, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, errors.New("user not found")
	}

	return GetUserByID(db, id)
}
func UpdatePassword(db *sql.DB, id int64, newRawPassword string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newRawPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password failed: %w", err)
	}

	query := `UPDATE users SET password = ? WHERE id = ?`
	res, err := db.Exec(query, string(hashedBytes), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func DeleteUser(db *sql.DB, id int64) error {
	query := `DELETE FROM users WHERE id = ?`

	res, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
