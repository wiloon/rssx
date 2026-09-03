package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"rssx/common"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/news"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
)

// testSecurityKey is the JWT signing key used across all e2e tests.
const testSecurityKey = "e2e-test-security-key-rssx"

// miniRedis backs redisx for the whole e2e run; resetState() flushes it per test.
var miniRedis *miniredis.Miniredis

func TestMain(m *testing.M) {
	// Set security key so JWT signing/parsing uses a known key during tests.
	os.Setenv("SECURITY_KEY", testSecurityKey)

	// Point redisx at an in-memory Redis. REDIS_ADDRESS must be set before the
	// first redisx call, since the connection pool initialises lazily once.
	mr, err := miniredis.Run()
	if err != nil {
		panic("failed to start miniredis: " + err.Error())
	}
	miniRedis = mr
	os.Setenv("REDIS_ADDRESS", mr.Addr())

	// Replace the DB initialized by init() with an in-memory instance so tests
	// are fully isolated from any on-disk database.
	common.InitForTesting()

	code := m.Run()
	mr.Close()
	os.Exit(code)
}

// resetState gives a test a clean SQLite DB and a clean Redis.
func resetState(t *testing.T) {
	t.Helper()
	common.InitForTesting()
	miniRedis.FlushAll()
}

// seedArticle stores an article in Redis and adds it to a feed's news index,
// exactly as the RSS sync path does.
func seedArticle(feedID int64, id, title string, score int64) {
	n := news.News{
		Id:          id,
		FeedId:      feedID,
		Title:       title,
		Url:         "https://example.com/" + id,
		Description: "<p>body " + id + "</p>",
		PubDate:     "2026-01-01",
		Guid:        id,
		Score:       score,
	}
	n.Save()
	list.NewList(0, feed.Feed{Id: feedID}).AppendNews(score, id)
}

// doRaw fires a request and returns the status code and raw response body, for
// the feed/news endpoints that return bare JSON rather than the ShowData envelope.
func doRaw(t *testing.T, method, path string, body interface{}) (int, []byte) {
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
	return w.Code, w.Body.Bytes()
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

// --- Redis-backed reader endpoints ---------------------------------------------

type feedItem struct {
	Id    int64
	Title string
}

type articleItem struct {
	Id       string
	Title    string
	NextId   string
	ReadFlag bool
}

// TestReaderFlow exercises the whole unread-window path through Redis:
// subscribe -> unread count -> unread window -> open one article (marks read,
// advances the boundary) -> mark the page read.
func TestReaderFlow(t *testing.T) {
	resetState(t)

	// Subscribe to a feed.
	code, body := doRaw(t, http.MethodPost, "/feed", map[string]string{
		"url": "https://example.com/rss", "title": "Example",
	})
	if code != http.StatusCreated {
		t.Fatalf("POST /feed: status %d, body %s", code, body)
	}
	var created feedItem
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatalf("POST /feed unmarshal: %v (%s)", err, body)
	}
	feedID := created.Id
	id := strconv.FormatInt(feedID, 10)

	// Seed three articles (oldest score first), as the sync path would.
	seedArticle(feedID, "a1", "Article One", 100)
	seedArticle(feedID, "a2", "Article Two", 200)
	seedArticle(feedID, "a3", "Article Three", 300)

	// GET /feeds: the feed carries "- 3" (three unread).
	titleOf := func(want int64) string {
		_, b := doRaw(t, http.MethodGet, "/feeds", nil)
		var feeds []feedItem
		if err := json.Unmarshal(b, &feeds); err != nil {
			t.Fatalf("GET /feeds unmarshal: %v (%s)", err, b)
		}
		for _, f := range feeds {
			if f.Id == want {
				return f.Title
			}
		}
		t.Fatalf("feed %d not in /feeds: %s", want, b)
		return ""
	}
	if got := titleOf(feedID); !strings.HasSuffix(got, " - 3") {
		t.Errorf("subscribed feed title = %q, want it to end with \" - 3\"", got)
	}

	// GET /news-list: the unread window has all three, none read.
	_, body = doRaw(t, http.MethodGet, "/news-list?id="+id, nil)
	var articles []articleItem
	if err := json.Unmarshal(body, &articles); err != nil {
		t.Fatalf("GET /news-list unmarshal: %v (%s)", err, body)
	}
	if len(articles) != 3 {
		t.Fatalf("unread window = %d articles, want 3: %s", len(articles), body)
	}
	for _, a := range articles {
		if a.ReadFlag {
			t.Errorf("article %s should be unread in a fresh window", a.Id)
		}
	}

	// GET /news for a1: returns its content and the next id, and marks it read.
	_, body = doRaw(t, http.MethodGet, "/news?feedId="+id+"&id=a1", nil)
	var one articleItem
	if err := json.Unmarshal(body, &one); err != nil {
		t.Fatalf("GET /news unmarshal: %v (%s)", err, body)
	}
	if one.Title != "Article One" {
		t.Errorf("GET /news Title = %q, want Article One", one.Title)
	}
	if one.NextId != "a2" {
		t.Errorf("GET /news NextId = %q, want a2", one.NextId)
	}

	// The read boundary advanced past a1: the window is now [a2, a3] and the
	// feed shows two unread.
	_, body = doRaw(t, http.MethodGet, "/news-list?id="+id, nil)
	json.Unmarshal(body, &articles)
	ids := make([]string, len(articles))
	for i, a := range articles {
		ids[i] = a.Id
	}
	if strings.Join(ids, ",") != "a2,a3" {
		t.Errorf("window after reading a1 = %v, want [a2 a3]", ids)
	}
	if got := titleOf(feedID); !strings.HasSuffix(got, " - 2") {
		t.Errorf("after reading a1, feed title = %q, want it to end with \" - 2\"", got)
	}

	// GET /mark-read: advance the boundary past the whole page; feed hits zero.
	doRaw(t, http.MethodGet, "/mark-read?feedId="+id, nil)
	if got := titleOf(feedID); !strings.HasSuffix(got, " - 0") {
		t.Errorf("after mark-read, feed title = %q, want it to end with \" - 0\"", got)
	}
}

// TestReaderFlow_UnknownFeed verifies an empty feed just yields an empty window.
func TestReaderFlow_UnknownFeed(t *testing.T) {
	resetState(t)

	code, body := doRaw(t, http.MethodGet, "/news-list?id=999", nil)
	if code != http.StatusOK {
		t.Fatalf("GET /news-list for empty feed: status %d", code)
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed != "null" && trimmed != "[]" {
		t.Errorf("empty feed window = %s, want null or []", trimmed)
	}
}
