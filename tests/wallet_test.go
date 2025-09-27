package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lupppig/stream-ledger-api/repository/kafka"
	"github.com/lupppig/stream-ledger-api/router"
)

type loginResponse struct {
	Data struct {
		AccessToken struct {
			Token string `json:"token"`
		} `json:"access_token"`
	} `json:"data"`
}

type walletResponse struct {
	Data struct {
		Wallet struct {
			WalletID int64  `json:"wallet_id"`
			Balance  int64  `json:"balance"`
			Currency string `json:"currency"`
		} `json:"wallet"`
	} `json:"data"`
}

func createAndLoginUser(router http.Handler, t *testing.T) string {
	signup := `{"first_name":"Wally","last_name":"Tester","email":"wallet@example.com","password":"secret"}`
	reqSignup, _ := http.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(signup))
	reqSignup.Header.Set("Content-Type", "application/json")
	rrSignup := httptest.NewRecorder()
	router.ServeHTTP(rrSignup, reqSignup)

	if rrSignup.Code != http.StatusCreated {
		t.Fatalf("failed to create test user, got %d\n", rrSignup.Code)
	}

	login := `{"email":"wallet@example.com","password":"secret"}`
	reqLogin, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(login))
	reqLogin.Header.Set("Content-Type", "application/json")
	rrLogin := httptest.NewRecorder()
	router.ServeHTTP(rrLogin, reqLogin)

	if rrLogin.Code != http.StatusOK {
		t.Fatalf("failed to login test user, got %d\n", rrLogin.Code)
	}

	var lr loginResponse
	if err := json.Unmarshal(rrLogin.Body.Bytes(), &lr); err != nil {
		t.Logf("%s", rrLogin.Body.String())
		t.Fatalf("failed to decode login response: %v\n", err)
	}

	if lr.Data.AccessToken.Token == "" {
		t.Fatal("expected access token, got empty string\n")
	}

	return lr.Data.AccessToken.Token
}

func TestGetWallet(t *testing.T) {
	pdb, mockProducer := SetupTestDB(t)

	prod := &kafka.Producer{Prod: mockProducer, Topic: "transaction"}
	router := router.Router(pdb, prod)

	token := createAndLoginUser(router, t)

	req, _ := http.NewRequest("GET", "/api/v1/wallet", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var wr walletResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &wr); err != nil {
		t.Fatalf("failed to decode wallet response: %v\n", err)
	}

	if wr.Data.Wallet.Currency != "NGN" {
		t.Errorf("expected NGN as default currency, got %s\n", wr.Data.Wallet.Currency)
	}
}
