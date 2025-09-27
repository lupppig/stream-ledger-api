package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lupppig/stream-ledger-api/controller"
	"github.com/lupppig/stream-ledger-api/controller/middleware"
	"github.com/lupppig/stream-ledger-api/repository/postgres"
)

func Router(db *postgres.PostgresDB) *mux.Router {
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

	return subr
}
