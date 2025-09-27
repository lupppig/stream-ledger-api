package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lupppig/stream-ledger-api/router"
)

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

type walletResponse struct {
	WalletID int64  `json:"wallet_id"`
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
}

func createAndLoginUser(router http.Handler, t *testing.T) string {
	signup := `{"first_name":"Wally","last_name":"Tester","email":"wallet@example.com","password":"secret"}`
	reqSignup, _ := http.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(signup))
	reqSignup.Header.Set("Content-Type", "application/json")
	rrSignup := httptest.NewRecorder()
	router.ServeHTTP(rrSignup, reqSignup)

	if rrSignup.Code != http.StatusCreated {
		t.Fatalf("failed to create test user, got %d", rrSignup.Code)
	}

	login := `{"email":"wallet@example.com","password":"secret"}`
	reqLogin, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(login))
	reqLogin.Header.Set("Content-Type", "application/json")
	rrLogin := httptest.NewRecorder()
	router.ServeHTTP(rrLogin, reqLogin)

	if rrLogin.Code != http.StatusOK {
		t.Fatalf("failed to login test user, got %d", rrLogin.Code)
	}

	var lr loginResponse
	if err := json.Unmarshal(rrLogin.Body.Bytes(), &lr); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	if lr.AccessToken == "" {
		t.Fatal("expected access token, got empty string")
	}

	return lr.AccessToken
}

func TestGetWallet(t *testing.T) {
	router := router.Router(pdb)

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
		t.Fatalf("failed to decode wallet response: %v", err)
	}

	if wr.Currency != "NGN" {
		t.Errorf("expected NGN as default currency, got %s", wr.Currency)
	}
}
