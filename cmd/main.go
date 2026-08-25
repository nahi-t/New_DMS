package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/docmanage_new/internal/user"
	_ "modernc.org/sqlite"
)

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

	mux := http.NewServeMux()

	user.SetupUserRoutes(mux, userHandler)

	fmt.Println("connected")

	err = http.ListenAndServe(":3000", mux)

	if err != nil {
		return

	}

}
