package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/lupppig/stream-ledger-api/router"
)

func TestConcurrentSignup(t *testing.T) {
	router := router.Router(pdb)

	var wg sync.WaitGroup
	numUsers := 5
	errors := make(chan error, numUsers)

	for i := 0; i < numUsers; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			payload := `{"first_name":"User","last_name":"` + string(rune('A'+i)) + `","email":"user` +
				string(rune('A'+i)) + `@example.com","password":"secret"}`

			req, _ := http.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusCreated {
				errors <- fmt.Errorf("expected %d, got %d for user%d", http.StatusCreated, rr.Code, i)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func signupTestUser(router http.Handler, t *testing.T) {
	payload := `{"first_name":"Jane","last_name":"Doe","email":"jane@example.com","password":"secret"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create test user, got status %d", rr.Code)
	}
}

func TestLogin(t *testing.T) {
	router := router.Router(pdb)

	signupTestUser(router, t)

	loginPayload := `{"email":"jane@example.com","password":"secret"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(loginPayload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected %d for valid login, got %d", http.StatusOK, rr.Code)
	}

	badPayload := `{"email":"jane@example.com","password":"wrongpass"}`
	reqBad, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(badPayload))
	reqBad.Header.Set("Content-Type", "application/json")

	rrBad := httptest.NewRecorder()
	router.ServeHTTP(rrBad, reqBad)

	if rrBad.Code == http.StatusOK {
		t.Errorf("expected login failure for wrong password, but got %d", rrBad.Code)
	}
}
