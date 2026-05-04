package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"rssx/common"

	"github.com/golang-jwt/jwt/v5"
)

// testSecurityKey is the JWT signing key used across all e2e tests.
const testSecurityKey = "e2e-test-security-key-rssx"

func TestMain(m *testing.M) {
	// Set security key so JWT signing/parsing uses a known key during tests.
	os.Setenv("SECURITY_KEY", testSecurityKey)

	// Replace the DB initialized by init() with an in-memory instance so tests
	// are fully isolated from any on-disk database.
	common.InitForTesting()

	os.Exit(m.Run())
}

// apiResponse mirrors the envelope returned by response.ShowData / ShowError.
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// doRequest fires an HTTP request against the test router and returns the parsed envelope.
func doRequest(t *testing.T, method, path string, body interface{}) (int, apiResponse) {
	t.Helper()
	router := setupRouter()

	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body %q: %v", w.Body.String(), err)
	}
	return w.Code, resp
}

// TestPing verifies the health-check endpoint.
func TestPing(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "pong") {
		t.Fatalf("expected body to contain 'pong', got: %s", w.Body.String())
	}
}

// TestRegister_Success verifies that a new user can register and receives a JWT token.
func TestRegister_Success(t *testing.T) {
	payload := map[string]string{"name": "e2e_user_reg", "password": "secret123"}
	statusCode, resp := doRequest(t, http.MethodPost, "/register", payload)

	if statusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", statusCode)
	}
	if resp.Code != 20000 {
		t.Fatalf("expected code 20000, got %d (message: %s)", resp.Code, resp.Message)
	}

	var data map[string]string
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to parse data field: %v", err)
	}
	token := data["token"]
	if token == "" {
		t.Fatal("expected token in response, got empty string")
	}

	// Verify token is a well-formed JWT (three dot-separated parts).
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected JWT to have 3 parts, got %d: %s", len(parts), token)
	}
}

// TestRegister_DuplicateUser verifies that registering the same username twice is rejected.
func TestRegister_DuplicateUser(t *testing.T) {
	payload := map[string]string{"name": "e2e_dup_user", "password": "secret123"}

	// First registration must succeed.
	_, first := doRequest(t, http.MethodPost, "/register", payload)
	if first.Code != 20000 {
		t.Fatalf("first registration failed unexpectedly: code=%d msg=%s", first.Code, first.Message)
	}

	// Second registration of the same username must fail.
	_, second := doRequest(t, http.MethodPost, "/register", payload)
	if second.Code == 20000 {
		t.Fatal("expected duplicate registration to fail, but it succeeded")
	}
}

// TestLogin_Success verifies that a registered user can log in and receives a JWT token.
func TestLogin_Success(t *testing.T) {
	// Register a user first.
	reg := map[string]string{"name": "e2e_user_login", "password": "loginpass"}
	_, regResp := doRequest(t, http.MethodPost, "/register", reg)
	if regResp.Code != 20000 {
		t.Fatalf("setup: registration failed: code=%d", regResp.Code)
	}

	// Now log in.
	statusCode, resp := doRequest(t, http.MethodPost, "/login", reg)
	if statusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", statusCode)
	}
	if resp.Code != 20000 {
		t.Fatalf("expected code 20000, got %d (message: %s)", resp.Code, resp.Message)
	}

	var data map[string]string
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to parse data field: %v", err)
	}
	if data["token"] == "" {
		t.Fatal("expected token in login response, got empty string")
	}
}

// TestLogin_WrongPassword verifies that incorrect credentials are rejected.
func TestLogin_WrongPassword(t *testing.T) {
	// Register a user first.
	reg := map[string]string{"name": "e2e_user_badpw", "password": "correctpass"}
	_, regResp := doRequest(t, http.MethodPost, "/register", reg)
	if regResp.Code != 20000 {
		t.Fatalf("setup: registration failed: code=%d", regResp.Code)
	}

	// Attempt login with wrong password.
	login := map[string]string{"name": "e2e_user_badpw", "password": "wrongpass"}
	_, resp := doRequest(t, http.MethodPost, "/login", login)
	if resp.Code == 20000 {
		t.Fatal("expected login with wrong password to fail, but it succeeded")
	}
}

// TestLogin_UnknownUser verifies that logging in as a non-existent user is rejected.
func TestLogin_UnknownUser(t *testing.T) {
	login := map[string]string{"name": "nobody_e2e", "password": "pass"}
	_, resp := doRequest(t, http.MethodPost, "/login", login)
	if resp.Code == 20000 {
		t.Fatal("expected login for unknown user to fail, but it succeeded")
	}
}

// TestLogin_TokenIsValidJWT verifies that the token returned by /login is a parseable JWT
// and contains the expected claims structure (iss, sub, aud, exp).
func TestLogin_TokenIsValidJWT(t *testing.T) {
	// Register then log in.
	creds := map[string]string{"name": "e2e_jwt_check", "password": "jwtpass"}
	doRequest(t, http.MethodPost, "/register", creds)

	_, resp := doRequest(t, http.MethodPost, "/login", creds)
	if resp.Code != 20000 {
		t.Fatalf("login failed: code=%d", resp.Code)
	}

	var data map[string]string
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to parse response data: %v", err)
	}
	tokenStr := data["token"]

	// Parse without verification to inspect claims.
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("token is not a valid JWT: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to read claims from token")
	}

	if claims["iss"] != "wiloon.com" {
		t.Errorf("expected iss=wiloon.com, got %v", claims["iss"])
	}
	if claims["sub"] != "rssx" {
		t.Errorf("expected sub=rssx, got %v", claims["sub"])
	}
	if claims["exp"] == nil {
		t.Error("expected exp claim to be set")
	}
	if claims["id"] == nil || claims["id"] == "" {
		t.Error("expected id claim to be set")
	}
}
