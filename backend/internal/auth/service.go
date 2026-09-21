package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-super-secret-key-change-this")

type Claims struct {
	UserID   int64  `json:"user_id"`
	Role     string `json:"role"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AuthUser struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthService struct {
	DB *pgxpool.Pool
}

func NewAuthService(db *pgxpool.Pool) *AuthService {
	return &AuthService{DB: db}
}

func (s *AuthService) Login(email, password string) (*AuthUser, string, error) {
	log.Printf("[DEBUG 1] Login attempt initiated for email: %s", email)

	// FIX 1: Changed '?' to '$1' for PostgreSQL pgx parameter binding
	query := `SELECT id, username, email, password, role, created_at FROM users WHERE email = $1`

	var u AuthUser
	var hashedPassword string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.DB.QueryRow(ctx, query, email).Scan(&u.ID, &u.Username, &u.Email, &hashedPassword, &u.Role, &u.CreatedAt)
	if err != nil {
		log.Printf("[DEBUG 2] Database lookup failed for %s: %v", email, err)

		// FIX 2: Replaced sql.ErrNoRows with pgx.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", errors.New("invalid email or password")
		}
		return nil, "", fmt.Errorf("db lookup failed: %w", err)
	}

	log.Printf("DEBUG: Found user ID %d for email %s", u.ID, u.Email)
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		log.Printf("[DEBUG 4] Password verification failed for email: %s", email)
		return nil, "", errors.New("invalid email or password")
	}

	token, err := s.generateToken(u.ID, u.Role, u.Username)
	if err != nil {
		return nil, "", fmt.Errorf("token signing failed: %w", err)
	}

	return &u, token, nil
}

func (s *AuthService) generateToken(userID int64, role string, username string) (string, error) {
	claims := &Claims{
		Username: username,
		UserID:   userID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
