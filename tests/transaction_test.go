package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/lupppig/stream-ledger-api/router"
)

func TestConcurrentTransactions(t *testing.T) {
	router := router.Router(pdb)

	token := createAndLoginUser(router, t)

	var wg sync.WaitGroup
	numRequests := 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			var payload string
			if i%2 == 0 {
				payload = `{"entry":"credit","amount":100}`
			} else {
				payload = `{"entry":"debit","amount":50}`
			}

			req, _ := http.NewRequest("POST", "/api/v1/transactions", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			results <- rr.Code
		}(i)
	}

	wg.Wait()
	close(results)

	var success, failure int
	for code := range results {
		if code == http.StatusCreated {
			success++
		} else {
			failure++
		}
	}

	if success == 0 {
		t.Fatal("expected at least one successful transaction, got none")
	}

	reqWallet, _ := http.NewRequest("GET", "/api/v1/wallet", nil)
	reqWallet.Header.Set("Authorization", "Bearer "+token)

	rrWallet := httptest.NewRecorder()
	router.ServeHTTP(rrWallet, reqWallet)

	if rrWallet.Code != http.StatusOK {
		t.Fatalf("expected 200 when fetching wallet, got %d", rrWallet.Code)
	}

	var wr walletResponse
	if err := json.Unmarshal(rrWallet.Body.Bytes(), &wr); err != nil {
		t.Fatalf("failed to decode wallet response: %v", err)
	}

	if wr.Balance < 0 {
		t.Fatalf("wallet balance went negative: %d", wr.Balance)
	}
}

type transactionsResponse struct {
	Transactions []struct {
		TransactionID int64  `json:"transaction_id"`
		Entry         string `json:"entry"`
		Amount        int64  `json:"amount"`
	} `json:"transactions"`
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
}

func seedTransactions(router http.Handler, token string, t *testing.T) {
	credits := []string{
		`{"entry":"credit","amount":200}`,
		`{"entry":"credit","amount":300}`,
	}
	debits := []string{
		`{"entry":"debit","amount":100}`,
	}

	payloads := append(credits, debits...)
	for _, p := range payloads {
		req, _ := http.NewRequest("POST", "/api/v1/transactions", strings.NewReader(p))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("failed to create transaction, got %d", rr.Code)
		}
	}
}

func TestGetTransactions(t *testing.T) {
	router := router.Router(pdb)

	token := createAndLoginUser(router, t)

	seedTransactions(router, token, t)

	req, _ := http.NewRequest("GET", "/api/v1/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	var resp transactionsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode transactions response: %v", err)
	}

	if len(resp.Transactions) == 0 {
		t.Fatal("expected at least one transaction, got none")
	}

	foundCredit, foundDebit := false, false
	for _, tx := range resp.Transactions {
		if tx.Entry == "credit" {
			foundCredit = true
		}
		if tx.Entry == "debit" {
			foundDebit = true
		}
	}

	if !foundCredit {
		t.Error("expected at least one credit transaction, found none")
	}
	if !foundDebit {
		t.Error("expected at least one debit transaction, found none")
	}
}


