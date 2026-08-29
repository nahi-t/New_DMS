package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func InitUserTable(db *pgxpool.Pool) error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id BIGSERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin', 'manager', 'user')),
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Exec(ctx, query)
	return err
}

func RegisterUser(db *pgxpool.Pool, username, email, rawPassword, role string) (*User, error) {
	if role == "" {
		role = RoleUser
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password failed: %w", err)
	}
	createdAt := time.Now().UTC()

	query := `INSERT INTO users (username, email, password, role, created_at) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int64
	err = db.QueryRow(ctx, query, username, email, string(hashedBytes), role, createdAt).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("user insertion failed: %w", err)
	}

	return &User{
		ID:        id,
		Username:  username,
		Email:     email,
		Role:      role,
		CreatedAt: createdAt,
	}, nil
}

func GetUserByID(db *pgxpool.Pool, id int64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, username, email, role, created_at FROM users WHERE id = $1`

	var u User
	err := db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return &u, nil
}

// UpdateProfile allows a regular user to update only their own username and email.
// The role is excluded from this query to prevent role escalation.
func UpdateProfile(db *pgxpool.Pool, id int64, username, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET username = $1, email = $2 WHERE id = $3`

	res, err := db.Exec(ctx, query, username, email, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	if res.RowsAffected() == 0 {
		return nil, errors.New("user not found")
	}

	return GetUserByID(db, id)
}

// AdminUpdateUser allows an administrator to update username, email, and role for any user.
func AdminUpdateUser(db *pgxpool.Pool, id int64, username, email, role string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET username = $1, email = $2, role = $3 WHERE id = $4`

	res, err := db.Exec(ctx, query, username, email, role, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if res.RowsAffected() == 0 {
		return nil, errors.New("user not found")
	}

	return GetUserByID(db, id)
}

func UpdatePassword(db *pgxpool.Pool, id int64, newRawPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newRawPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password failed: %w", err)
	}

	query := `UPDATE users SET password = $1 WHERE id = $2`
	res, err := db.Exec(ctx, query, string(hashedBytes), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if res.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func DeleteUser(db *pgxpool.Pool, id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM users WHERE id = $1`

	res, err := db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if res.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func GetAllUsers(db *pgxpool.Pool) ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, username, email, role, created_at FROM users ORDER BY id ASC`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
