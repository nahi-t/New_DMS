package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	docs "github.com/docmanage_new/cmd/docs"
	auditlog "github.com/docmanage_new/internal/auditLog"
	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/document"
	"github.com/docmanage_new/internal/documentversion"
	"github.com/docmanage_new/internal/folder"
	"github.com/docmanage_new/internal/share"
	"github.com/docmanage_new/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
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
	/*  Audit                                                          */
	/* -------------------------------------------------------------- */

	auditLogRepo := auditlog.NewRepository(db)
	auditLogService := auditlog.NewService(auditLogRepo)
	auditHandler := auditlog.NewAuditHandler(auditLogService)

	/* -------------------------------------------------------------- */
	/*  User + auth                                                    */
	/* -------------------------------------------------------------- */

	userHandler := user.NewUserHandler(db)
	authservice := auth.NewAuthService(db)
	authHandler := auth.NewAuthHandler(authservice, auditLogService)
	authMW := auth.NewMiddleware(authservice)

	/* -------------------------------------------------------------- */
	/*  Share                                                          */
	/* -------------------------------------------------------------- */

	shareRepo := share.NewRepository(db)
	shareService := share.NewService(shareRepo)
	shareHandler := share.NewHandler(shareService, auditLogService)

	/* -------------------------------------------------------------- */
	/*  Document version — repo + storage + service + handler          */
	/* -------------------------------------------------------------- */

	versionRepo := documentversion.NewRepository(db)

	storageMgr, err := documentversion.NewStorageManager("documents")
	if err != nil {
		log.Fatalf("failed to init Cloudinary storage: %v", err)
	}

	versionSvc := documentversion.NewService(versionRepo, storageMgr)
	versionHandler := documentversion.NewDocumentVersionHandler(versionSvc)

	/* -------------------------------------------------------------- */
	/*  Document — handler depends on the version service + shares     */
	/* -------------------------------------------------------------- */

	docHandler := document.NewHandler(db, versionSvc, auditLogService, shareService)

	/* -------------------------------------------------------------- */
	/*  Folder                                                         */
	/* -------------------------------------------------------------- */

	folderRepo := folder.NewRepository(db)
	folderService := folder.NewService(folderRepo)
	folderHandler := folder.NewHandler(folderService)

	/* -------------------------------------------------------------- */
	/*  Mux                                                            */
	/* -------------------------------------------------------------- */

	mux := http.NewServeMux()

	/* -------------------------------------------------------------- */
	/*  Swagger — must be registered AFTER mux is created               */
	/* -------------------------------------------------------------- */

	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Host = "" // empty = use the same host as the browser

	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(docs.SwaggerInfo.ReadDoc()))
	})

	/* -------------------------------------------------------------- */
	/*  Application routes                                             */
	/* -------------------------------------------------------------- */

	folder.SetupFolderRoutes(mux, folderHandler, authMW)
	user.SetupUserRoutes(mux, userHandler, authMW)
	auth.SetupAuthRoutes(mux, authHandler)
	document.RegisterRoutes(mux, docHandler, authMW)
	documentversion.RegisterDocumentVersionRoutes(mux, versionHandler, authMW)
	share.RegisterRoutes(mux, shareHandler, authMW)
	auditlog.SetUpAuditLogDisplay(mux, auditHandler, authMW)

	/* -------------------------------------------------------------- */
	/*  Server                                                         */
	/* -------------------------------------------------------------- */

	handlerWithCORS := enableCORS(mux)

	log.Println("Swagger UI:   http://localhost:8080/swagger/index.html")
	log.Println("Server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
