package main

import (
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
	"github.com/lupppig/stream-ledger-api/model/migrations"
	"github.com/lupppig/stream-ledger-api/repository/postgres"
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

	// run database migrations
	if err := migrations.RunMigrations(db.DB); err != nil {
		log.Printf("failed to perform migrations: %v", err)
	}

	router := mux.NewRouter()
	subr := router.PathPrefix("/api/v1").Subrouter()

	c := controller.Router{DB: db}
	// authentication routes
	subr.HandleFunc("/auth/signup", c.RegisterUser).Methods("POST")
	subr.HandleFunc("/auth/login", c.SignIn).Methods("POST")

	// transaction routes & wallet routes
	subr.Handle("/wallet", middleware.AuthMiddleware(http.HandlerFunc(c.GetWallet))).Methods("GET")
	subr.Handle("/transactions", middleware.AuthMiddleware(http.HandlerFunc(c.CreateTransactions))).Methods("POST")

	srv := &http.Server{
		Handler:      subr,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	// will handle graceful shutdown
	log.Fatal(srv.ListenAndServe())
}
