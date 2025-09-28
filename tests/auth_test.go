package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lupppig/stream-ledger-api/repository/kafka"
	"github.com/lupppig/stream-ledger-api/router"
)

func TestSignup(t *testing.T) {

	pdb, mockProducer := SetupTestDB(t)
	prod := &kafka.Producer{Prod: mockProducer, Topic: "transaction"}
	r := router.Router(pdb, prod)

	numUsers := 5
	for i := 0; i < numUsers; i++ {
		payload := fmt.Sprintf(
			`{"first_name":"User","last_name":"%c","email":"user%c@example.com","password":"secret"}`,
			'A'+i, 'A'+i,
		)

		req, _ := http.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d for user %d", http.StatusCreated, rr.Code, i)
		}
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	// Initialize router
	pdb, mockProducer := SetupTestDB(t)
	prod := &kafka.Producer{Prod: mockProducer, Topic: "transaction"}
	router := router.Router(pdb, prod)

	user := map[string]string{
		"email":      "duplicate@example.com",
		"password":   "StrongPass123",
		"first_name": "Alice",
		"last_name":  "Smith",
	}

	body, _ := json.Marshal(user)

	req1 := httptest.NewRequest("POST", "/api/v1/auth/signup", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	resp1 := httptest.NewRecorder()
	router.ServeHTTP(resp1, req1)

	if resp1.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp1.Code)
	}

	req2 := httptest.NewRequest("POST", "/api/v1/auth/signup", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)

	if resp2.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", resp2.Code)
	}

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	msg, ok := respBody["message"].(string)
	if !ok {
		t.Fatalf("response body missing 'message' field")
	}

	if msg != "email already exists" {
		t.Fatalf("expected error message 'email already exists', got '%s'", msg)
	}
}

func TestLogin(t *testing.T) {
	pdb, mockProducer := SetupTestDB(t)

	prod := &kafka.Producer{Prod: mockProducer, Topic: "transaction"}
	r := router.Router(pdb, prod)

	numUsers := 5
	// create users first
	for i := 0; i < numUsers; i++ {
		payload := fmt.Sprintf(
			`{"first_name":"User","last_name":"%c","email":"loginuser%c@example.com","password":"secret"}`,
			'A'+i, 'A'+i,
		)

		req, _ := http.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("failed to create test user %d, got %d", i, rr.Code)
		}
	}

	// now login each user
	for i := 0; i < numUsers; i++ {
		loginPayload := fmt.Sprintf(
			`{"email":"loginuser%c@example.com","password":"secret"}`,
			'A'+i,
		)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(loginPayload))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected %d for valid login of user %d, got %d", http.StatusOK, i, rr.Code)
		}
	}

	// test a bad login for one user
	badPayload := `{"email":"loginuserA@example.com","password":"wrongpass"}`
	reqBad, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(badPayload))
	reqBad.Header.Set("Content-Type", "application/json")

	rrBad := httptest.NewRecorder()
	r.ServeHTTP(rrBad, reqBad)

	if rrBad.Code == http.StatusOK {
		t.Errorf("expected login failure for wrong password, but got %d", rrBad.Code)
	}
}
