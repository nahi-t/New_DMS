// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"time"

// 	"github.com/docmanage_new/internal/auth"
// 	"github.com/docmanage_new/internal/document"
// 	"github.com/docmanage_new/internal/documentversion"
// 	"github.com/docmanage_new/internal/folder"
// 	"github.com/docmanage_new/internal/user"
// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/joho/godotenv"
// )

// func enableCORS(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

// 		if r.Method == http.MethodOptions {
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// func OpenDB() (*pgxpool.Pool, error) {
// 	_ = godotenv.Load()
// 	connStr := os.Getenv("DATABASE_URL")
// 	if connStr == "" {
// 		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
// 	}

// 	ctx := context.Background()

// 	// 1. Parse connection string into pgxpool Config struct
// 	config, err := pgxpool.ParseConfig(connStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to parse connection string: %w", err)
// 	}

// 	// 2. CRITICAL FOR SUPABASE: Disable prepared statement caching
// 	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

// 	// 3. Connect using configured pool settings
// 	pool, err := pgxpool.NewWithConfig(ctx, config)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to create connection pool: %w", err)
// 	}

// 	if err := pool.Ping(ctx); err != nil {
// 		pool.Close()
// 		return nil, fmt.Errorf("database ping failed: %w", err)
// 	}

// 	var version string
// 	if err := pool.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
// 		pool.Close()
// 		return nil, fmt.Errorf("version query failed: %w", err)
// 	}

// 	log.Println("Connected to PostgreSQL:", version)
// 	return pool, nil
// }

// func main() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	db, err := OpenDB()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	fmt.Println("db connected")

// 	if err := user.InitUserTable(db); err != nil {
// 		log.Fatalf("Failed to initialize user table: %v", err)
// 	}
// 	fmt.Println("User table initialized.")

// 	// Admin check/seeding: Logs status and proceeds naturally without crashing
// 	msg := user.SeedAdminUser(ctx, db)
// 	log.Println(msg)

// 	userHandler := user.NewUserHandler(db)
// 	authservice := auth.NewAuthService(db)
// 	docHandler := document.NewHandler(db)

// 	documentVersionService := documentversion.NewService(documentversion.NewRepository(db))
// 	documentvesionHandler := documentversion.NewDocumentVersionHandler(documentVersionService)
// 	documentversion.NewStorageManager("./versions") // Ensure the storage manager is initialized with a base directory
// 	authHandler := auth.NewAuthHandler(authservice)
// 	authMiddleware := auth.NewMiddleware(authservice)

// 	folderRepo := folder.NewRepository(db)
// 	folderService := folder.NewService(folderRepo)
// 	folderHandler := folder.NewHandler(folderService)

// 	mux := http.NewServeMux()
// 	handlerWithCORS := enableCORS(mux)

// 	folder.SetupFolderRoutes(mux, folderHandler, authMiddleware)
// 	user.SetupUserRoutes(mux, userHandler, authMiddleware)
// 	auth.SetupAuthRoutes(mux, authHandler)
// 	document.RegisterRoutes(mux, docHandler, authMiddleware)
// 	documentversion.RegisterDocumentVersionRoutes(mux, documentvesionHandler, authMiddleware)

// 	log.Println("Server running on http://localhost:8080")
// 	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
// 		log.Fatalf("Server failed to start: %v", err)
// 	}
// }

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	auditlog "github.com/docmanage_new/internal/auditLog"
	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/document"
	"github.com/docmanage_new/internal/documentversion"
	"github.com/docmanage_new/internal/folder"
	"github.com/docmanage_new/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func OpenDB() (*pgxpool.Pool, error) {
	_ = godotenv.Load()
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Disable prepared statement caching (for Supabase / Neon poolers).
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	var version string
	if err := pool.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		pool.Close()
		return nil, fmt.Errorf("version query failed: %w", err)
	}

	log.Println("Connected to PostgreSQL:", version)
	return pool, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	/* -------------------------------------------------------------- */
	/*  Database                                                       */
	/* -------------------------------------------------------------- */

	db, err := OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("db connected")

	/* -------------------------------------------------------------- */
	/*  Bootstrap user table + admin                                   */
	/* -------------------------------------------------------------- */

	if err := user.InitUserTable(db); err != nil {
		log.Fatalf("Failed to initialize user table: %v", err)
	}
	fmt.Println("User table initialized.")

	msg := user.SeedAdminUser(ctx, db)
	log.Println(msg)

	/* -------------------------------------------------------------- */
	/*  User + auth                                                    */
	/* -------------------------------------------------------------- */

	userHandler := user.NewUserHandler(db)
	authservice := auth.NewAuthService(db)
	authHandler := auth.NewAuthHandler(authservice)
	authMW := auth.NewMiddleware(authservice)

	/* -------------------------------------------------------------- */
	/*  Document version — repo + storage + service + handler          */
	/*   Order matters: StorageManager must be created BEFORE        */
	/*      the Service that depends on it.                            */
	/* -------------------------------------------------------------- */

	versionRepo := documentversion.NewRepository(db)

	// StorageManager now wraps Cloudinary (no local folder needed).
	// The argument is the top-level Cloudinary folder name.
	storageMgr, err := documentversion.NewStorageManager("documents")
	if err != nil {
		log.Fatalf("failed to init Cloudinary storage: %v", err)
	}

	versionSvc := documentversion.NewService(versionRepo, storageMgr)
	versionHandler := documentversion.NewDocumentVersionHandler(versionSvc)

	/* -------------------------------------------------------------- */
	/*  Document — handler depends on the version service              */
	/* -------------------------------------------------------------- */
	auditLogRepo := auditlog.NewRepository(db)
	auditLogService := auditlog.NewService(auditLogRepo)
	auditHandler := auditlog.NewAuditHandler(auditLogService)

	docHandler := document.NewHandler(db, versionSvc, auditLogService)

	/* -------------------------------------------------------------- */
	/*  Folder                                                         */
	/* -------------------------------------------------------------- */

	folderRepo := folder.NewRepository(db)
	folderService := folder.NewService(folderRepo)
	folderHandler := folder.NewHandler(folderService)

	/* -------------------------------------------------------------- */
	/*  Routes                                                         */
	/* -------------------------------------------------------------- */

	mux := http.NewServeMux()
	handlerWithCORS := enableCORS(mux)

	folder.SetupFolderRoutes(mux, folderHandler, authMW)
	user.SetupUserRoutes(mux, userHandler, authMW)
	auth.SetupAuthRoutes(mux, authHandler)
	document.RegisterRoutes(mux, docHandler, authMW)
	documentversion.RegisterDocumentVersionRoutes(mux, versionHandler, authMW)
	auditlog.SetUpAuditLogDisplay(mux, auditHandler, authMW)

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
