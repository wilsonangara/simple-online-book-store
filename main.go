package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kenshaw/envcfg"

	"github.com/wilsonangara/simple-online-book-store/auth"
	"github.com/wilsonangara/simple-online-book-store/handlers/user"
	"github.com/wilsonangara/simple-online-book-store/storage/sqlite"
	user_storage "github.com/wilsonangara/simple-online-book-store/storage/sqlite/user"
)

var config *envcfg.Envcfg

func init() {
	var err error
	config, err = envcfg.New()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	handlers := setupHandlers()

	addr := net.JoinHostPort("", config.GetString("server.port"))
	srv := &http.Server{
		Addr:    addr,
		Handler: handlers,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		const shutdownTimeout = 10 * time.Second

		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		ctx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("HTTP server shutdown: %v", err)
		}
		log.Print("HTTP server shutdown")
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed

	log.Print("server exited gracefully")
}

func setupHandlers() http.Handler {
	r := gin.Default()

	// handle default endpoint response
	r.Any("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{
			"message": "simple online book store server",
		})
	})

	// handle if page is not found
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, map[string]any{
			"message": "endpoint not found",
		})
	})

	authClient, err := auth.NewClient(config.GetString("jwt.secret"))
	if err != nil {
		log.Fatalf("failed to initialize auth client: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}
	dbName := filepath.Join(wd, "storage", "sqlite", "databases",
		config.GetString("db.name"),
	)
	pathToMigrations := filepath.Join(wd, "storage", "migrations")

	// storage
	storage, err := sqlite.NewStorage(dbName, pathToMigrations)
	if err != nil {
		log.Fatalf("failed to initialized storage: %v", err)
	}

	userStorage := user_storage.NewStorage(storage.Database())

	v1 := r.Group("/v1")

	userHandler := user.NewHandler(authClient, userStorage)
	userHandler.AddUserRoutes(v1)

	return r
}
