package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/lupppig/stream-ledger-api/controller"
	"github.com/lupppig/stream-ledger-api/controller/middleware"
	"github.com/lupppig/stream-ledger-api/jobs"
	"github.com/lupppig/stream-ledger-api/model/migrations"
	"github.com/lupppig/stream-ledger-api/repository/postgres"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/riverqueue/river/rivermigrate"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load env variables: %v", err)
	}

	connString := os.Getenv("DB_URL")
	portStr := os.Getenv("PORT")
	port, _ := strconv.Atoi(portStr)

	db, err := postgres.Connect(connString)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := setupRiver(db); err != nil {
		log.Printf("error setting up river: %v", err)
	}

	// run database migrations
	if err := migrations.RunMigrations(db.DB); err != nil {
		log.Printf("failed to perform migrations: %v", err)
	}

	router := mux.NewRouter()
	subr := router.PathPrefix("/api/v1").Subrouter()
	subr.Use(middleware.LoggingMiddleware)

	c := controller.Router{DB: db}
	// authentication routes
	subr.HandleFunc("/auth/signup", c.RegisterUser).Methods("POST")
	subr.HandleFunc("/auth/login", c.SignIn).Methods("POST")

	// transaction routes & wallet routes
	subr.Handle("/wallet", middleware.AuthMiddleware(http.HandlerFunc(c.GetWallet))).Methods("GET")
	subr.Handle("/transactions", middleware.AuthMiddleware(http.HandlerFunc(c.CreateTransactions))).Methods("POST")
	subr.Handle("/transactions", middleware.AuthMiddleware(http.HandlerFunc(c.ListUserTransactions))).Methods("GET")
	subr.Handle("/transactions/export", middleware.AuthMiddleware(http.HandlerFunc(c.ExportTransaction))).Methods("POST")

	srv := &http.Server{
		Handler:      subr,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	log.Printf("Server started on addr: %v", srv.Addr)
	// will handle graceful shutdown
	log.Fatal(srv.ListenAndServe())
}

func setupRiver(db *postgres.PostgresDB) error {

	migrator, err := rivermigrate.New(riverdatabasesql.New(db.DB.DB), nil)

	if err != nil {
		return err
	}
	_, err = migrator.Migrate(context.Background(), rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
	if err != nil {
		return fmt.Errorf("failed to run river migrations: %w", err)
	}
	workers := river.NewWorkers()
	river.AddWorker(workers, &jobs.ExportTransactionsWorker{
		DB: db,
	})

	riverClient, err := river.NewClient(riverdatabasesql.New(db.DB.DB), &river.Config{
		JobTimeout: 10 * time.Minute,
		Workers:    workers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
	})

	if err != nil {
		return err
	}

	db.River = riverClient

	if err := riverClient.Start(context.Background()); err != nil {
		return err
	}

	return nil
}
