package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/folder"
	"github.com/docmanage_new/internal/user"
	_ "modernc.org/sqlite"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from your frontend origin (or use "*" for all origins during development)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle browser preflight OPTIONS requests immediately
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "app.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ping verifies the database file is readable and reachable
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to reach database: %w", err)
	}

	return db, nil
}
func main() {
	db, err := OpenDB()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	fmt.Println("db connected")

	if err := user.InitUserTable(db); err != nil {
		log.Fatalf("Failed to initialize user table: %v", err)
	}
	fmt.Println("User table initialized.")

	userHandler := user.NewUserHandler(db)
	authservice := auth.NewAuthService(db)
	authHandler := auth.NewAuthHandler(authservice)
	authMiddleware := auth.NewMiddleware(authservice)

	folderRepo := folder.NewRepository(db)
	folderService := folder.NewService(folderRepo)
	folderHandler := folder.NewHandler(folderService)

	mux := http.NewServeMux()
	handlerWithCORS := enableCORS(mux)
	folder.SetupFolderRoutes(mux, folderHandler, authMiddleware)

	user.SetupUserRoutes(mux, userHandler, authMiddleware)
	auth.SetupAuthRoutes(mux, authHandler)

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
