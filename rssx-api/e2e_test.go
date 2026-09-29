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
	"time"

	"rssx/common"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/news"
	rssxjwt "rssx/utils/jwt"

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
	os.Setenv("RSSX_SECURITY_KEY", testSecurityKey)

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

// doRaw fires an authenticated request and returns the status code and raw
// response body, for the feed/news endpoints that return bare JSON rather than
// the ShowData envelope.
func doRaw(t *testing.T, method, path string, body interface{}) (int, []byte) {
	t.Helper()
	return doRawAs(t, testUserAlice, method, path, body)
}

// Test user ids; tokens only need an id claim, not a users row.
const (
	testUserAlice = "e2e-alice"
	testUserBob   = "e2e-bob"
)

// doRawAs is doRaw authenticated as the given user id.
func doRawAs(t *testing.T, userId, method, path string, body interface{}) (int, []byte) {
	t.Helper()
	return doRawWithAuth(t, method, path, body, "Bearer "+rssxjwt.NewToken(userId))
}

// doRawWithAuth is doRaw with an explicit Authorization header value; an empty
// value sends no header at all.
func doRawWithAuth(t *testing.T, method, path string, body interface{}, authorization string) (int, []byte) {
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
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

// TestProtectedEndpoints_RequireAuth verifies every reader and feed-management
// endpoint rejects requests without a valid bearer token.
func TestProtectedEndpoints_RequireAuth(t *testing.T) {
	resetState(t)

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  testUserAlice,
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	expiredToken, _ := expired.SignedString([]byte(testSecurityKey))
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  testUserAlice,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	forgedToken, _ := forged.SignedString([]byte("not-the-server-key"))
	noId := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	noIdToken, _ := noId.SignedString([]byte(testSecurityKey))

	authHeaders := map[string]string{
		"no header":         "",
		"malformed token":   "Bearer not-a-jwt",
		"expired token":     "Bearer " + expiredToken,
		"wrong signing key": "Bearer " + forgedToken,
		"token without id":  "Bearer " + noIdToken,
	}
	endpoints := []struct{ method, path string }{
		{http.MethodGet, "/feeds"},
		{http.MethodGet, "/feeds/detail"},
		{http.MethodPost, "/feed"},
		{http.MethodPut, "/feed/1"},
		{http.MethodDelete, "/feed/1"},
		{http.MethodDelete, "/feed/1/purge"},
		{http.MethodPost, "/sync"},
		{http.MethodPost, "/sync/1"},
		{http.MethodGet, "/news-list?id=1"},
		{http.MethodGet, "/news?feedId=1&id=x"},
		{http.MethodGet, "/previous-news?feedId=1&newsId=x"},
		{http.MethodGet, "/mark-read?feedId=1"},
	}
	for name, header := range authHeaders {
		for _, ep := range endpoints {
			code, body := doRawWithAuth(t, ep.method, ep.path, nil, header)
			if code != http.StatusUnauthorized {
				t.Errorf("%s %s with %s: status %d, want 401 (body %s)", ep.method, ep.path, name, code, body)
			}
		}
	}
}

// TestPublicEndpoints_NoAuth verifies ping, login and register stay reachable
// without a token.
func TestPublicEndpoints_NoAuth(t *testing.T) {
	resetState(t)

	if code, _ := doRawWithAuth(t, http.MethodGet, "/ping", nil, ""); code != http.StatusOK {
		t.Errorf("GET /ping without token: status %d, want 200", code)
	}
	creds := map[string]string{"name": "e2e_public", "password": "pw"}
	if code, _ := doRawWithAuth(t, http.MethodPost, "/register", creds, ""); code == http.StatusUnauthorized {
		t.Error("POST /register without token was rejected with 401")
	}
	if code, _ := doRawWithAuth(t, http.MethodPost, "/login", creds, ""); code == http.StatusUnauthorized {
		t.Error("POST /login without token was rejected with 401")
	}
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

// TestFeedManagementFlow exercises the feed maintenance endpoints:
// create -> detail list -> rename (PUT) -> purge (hard delete), and checks that
// the purge also wipes the feed's articles and index from Redis.
func TestFeedManagementFlow(t *testing.T) {
	resetState(t)

	// Create a feed.
	code, body := doRaw(t, http.MethodPost, "/feed", map[string]string{
		"url": "https://example.com/mgmt", "title": "Before",
	})
	if code != http.StatusCreated {
		t.Fatalf("POST /feed: status %d, body %s", code, body)
	}
	var created feedItem
	json.Unmarshal(body, &created)
	feedID := created.Id
	id := strconv.FormatInt(feedID, 10)

	seedArticle(feedID, "m1", "Item One", 100)
	seedArticle(feedID, "m2", "Item Two", 200)

	// GET /feeds/detail: the feed is listed with its raw title and url.
	code, body = doRaw(t, http.MethodGet, "/feeds/detail", nil)
	if code != http.StatusOK {
		t.Fatalf("GET /feeds/detail: status %d", code)
	}
	var detail []struct {
		Id    int64
		Title string
		Url   string
	}
	if err := json.Unmarshal(body, &detail); err != nil {
		t.Fatalf("GET /feeds/detail unmarshal: %v (%s)", err, body)
	}
	found := false
	for _, f := range detail {
		if f.Id == feedID {
			found = true
			if f.Title != "Before" || f.Url != "https://example.com/mgmt" {
				t.Errorf("detail row = %+v, want raw title/url", f)
			}
		}
	}
	if !found {
		t.Fatalf("feed %d missing from /feeds/detail: %s", feedID, body)
	}

	// PUT /feed/:id renames it.
	code, body = doRaw(t, http.MethodPut, "/feed/"+id, map[string]string{
		"url": "https://example.com/mgmt", "title": "After",
	})
	if code != http.StatusOK {
		t.Fatalf("PUT /feed/%s: status %d, body %s", id, code, body)
	}
	var updated feedItem
	json.Unmarshal(body, &updated)
	if updated.Title != "After" {
		t.Errorf("PUT /feed title = %q, want After", updated.Title)
	}

	// DELETE /feed/:id/purge removes it outright.
	code, body = doRaw(t, http.MethodDelete, "/feed/"+id+"/purge", nil)
	if code != http.StatusNoContent {
		t.Fatalf("DELETE /feed/%s/purge: status %d, body %s", id, code, body)
	}

	// The feed is gone from the detail list.
	_, body = doRaw(t, http.MethodGet, "/feeds/detail", nil)
	json.Unmarshal(body, &detail)
	for _, f := range detail {
		if f.Id == feedID {
			t.Errorf("feed %d still in /feeds/detail after purge: %s", feedID, body)
		}
	}

	// Redis: the article index is empty and the article hashes are gone.
	if n := list.Count(int(feedID)); n != 0 {
		t.Errorf("feed_news index size = %d after purge, want 0", n)
	}
	for _, k := range []string{"news:m1", "news:m2"} {
		if miniRedis.Exists(k) {
			t.Errorf("Redis key %q still exists after purge", k)
		}
	}

	// Purging a feed that does not exist is a 404.
	code, _ = doRaw(t, http.MethodDelete, "/feed/"+id+"/purge", nil)
	if code != http.StatusNotFound {
		t.Errorf("purge of missing feed: status %d, want 404", code)
	}
}

// TestPreviousAndMissingArticles covers the edge cases that used to panic or
// return the wrong article: no previous article, unknown ids, GC'd articles,
// and malformed feed ids.
func TestPreviousAndMissingArticles(t *testing.T) {
	resetState(t)
	seedArticle(5, "p1", "First", 100)
	seedArticle(5, "p2", "Second", 200)

	code, body := doRaw(t, http.MethodGet, "/previous-news?feedId=5&newsId=p2", nil)
	if code != http.StatusOK {
		t.Fatalf("previous of p2: status %d, body %s", code, body)
	}
	var prev articleItem
	json.Unmarshal(body, &prev)
	if prev.Id != "p1" || prev.NextId != "p2" {
		t.Errorf("previous of p2 = %+v, want p1 with next p2", prev)
	}

	cases := []struct {
		name, path string
		want       int
	}{
		{"previous of first article", "/previous-news?feedId=5&newsId=p1", http.StatusNotFound},
		{"previous of unknown article", "/previous-news?feedId=5&newsId=nope", http.StatusNotFound},
		{"previous in empty feed", "/previous-news?feedId=6&newsId=p1", http.StatusNotFound},
		{"previous with bad feedId", "/previous-news?feedId=abc&newsId=p2", http.StatusBadRequest},
		{"unknown article", "/news?feedId=5&id=nope", http.StatusNotFound},
		{"article with bad feedId", "/news?feedId=abc&id=p1", http.StatusBadRequest},
	}
	for _, tc := range cases {
		if code, body := doRaw(t, http.MethodGet, tc.path, nil); code != tc.want {
			t.Errorf("%s: status %d, want %d (body %s)", tc.name, code, tc.want, body)
		}
	}

	// An article still in the index but whose hash was removed (GC race).
	miniRedis.Del("news:p1")
	if code, body := doRaw(t, http.MethodGet, "/news?feedId=5&id=p1", nil); code != http.StatusNotFound {
		t.Errorf("article with missing hash: status %d, want 404 (body %s)", code, body)
	}
}

// feedIds returns the ids in GET /feeds/detail for one user.
func feedIdsOf(t *testing.T, userId string) map[int64]bool {
	t.Helper()
	code, body := doRawAs(t, userId, http.MethodGet, "/feeds/detail", nil)
	if code != http.StatusOK {
		t.Fatalf("GET /feeds/detail as %s: status %d", userId, code)
	}
	var feeds []feedItem
	if err := json.Unmarshal(body, &feeds); err != nil {
		t.Fatalf("GET /feeds/detail unmarshal: %v (%s)", err, body)
	}
	out := map[int64]bool{}
	for _, f := range feeds {
		out[f.Id] = true
	}
	return out
}

// unreadOf returns the "- N" unread suffix of one feed in GET /feeds for a user.
func unreadOf(t *testing.T, userId string, feedID int64) string {
	t.Helper()
	_, body := doRawAs(t, userId, http.MethodGet, "/feeds", nil)
	var feeds []feedItem
	json.Unmarshal(body, &feeds)
	for _, f := range feeds {
		if f.Id == feedID {
			return f.Title[strings.LastIndex(f.Title, " - ")+3:]
		}
	}
	t.Fatalf("feed %d not in /feeds for %s: %s", feedID, userId, body)
	return ""
}

func addFeedAs(t *testing.T, userId, url, title string) int64 {
	t.Helper()
	code, body := doRawAs(t, userId, http.MethodPost, "/feed", map[string]string{"url": url, "title": title})
	if code != http.StatusCreated {
		t.Fatalf("POST /feed as %s: status %d, body %s", userId, code, body)
	}
	var f feedItem
	json.Unmarshal(body, &f)
	return f.Id
}

// TestMultiUserIsolation checks that subscriptions and read state are per user
// while feeds and their articles are shared.
func TestMultiUserIsolation(t *testing.T) {
	resetState(t)

	shared := addFeedAs(t, testUserAlice, "https://example.com/shared", "Shared")
	aliceOnly := addFeedAs(t, testUserAlice, "https://example.com/alice", "Alice only")
	if again := addFeedAs(t, testUserBob, "https://example.com/shared", "Shared"); again != shared {
		t.Fatalf("same URL got feed id %d for bob, want shared id %d", again, shared)
	}
	sid := strconv.FormatInt(shared, 10)
	seedArticle(shared, "s1", "One", 100)
	seedArticle(shared, "s2", "Two", 200)

	// Subscriptions are per user.
	if got := feedIdsOf(t, testUserBob); len(got) != 1 || !got[shared] {
		t.Errorf("bob's feeds = %v, want only the shared feed %d", got, shared)
	}
	if got := feedIdsOf(t, testUserAlice); len(got) != 2 || !got[aliceOnly] {
		t.Errorf("alice's feeds = %v, want shared + alice-only", got)
	}

	// Read state is per user.
	doRawAs(t, testUserAlice, http.MethodGet, "/news?feedId="+sid+"&id=s1", nil)
	if got := unreadOf(t, testUserAlice, shared); got != "1" {
		t.Errorf("alice unread after reading s1 = %s, want 1", got)
	}
	if got := unreadOf(t, testUserBob, shared); got != "2" {
		t.Errorf("bob unread after alice read s1 = %s, want 2", got)
	}
	doRawAs(t, testUserBob, http.MethodGet, "/mark-read?feedId="+sid, nil)
	if got := unreadOf(t, testUserAlice, shared); got != "1" {
		t.Errorf("alice unread after bob marked all read = %s, want 1", got)
	}

	// Another user's feed cannot be edited, purged, or synced.
	aid := strconv.FormatInt(aliceOnly, 10)
	if code, _ := doRawAs(t, testUserBob, http.MethodPut, "/feed/"+aid, map[string]string{"url": "https://example.com/x", "title": "X"}); code != http.StatusNotFound {
		t.Errorf("bob editing alice's feed: status %d, want 404", code)
	}
	if code, _ := doRawAs(t, testUserBob, http.MethodDelete, "/feed/"+aid+"/purge", nil); code != http.StatusNotFound {
		t.Errorf("bob purging alice's feed: status %d, want 404", code)
	}
	if code, _ := doRawAs(t, testUserBob, http.MethodPost, "/sync/"+aid, nil); code != http.StatusNotFound {
		t.Errorf("bob syncing alice's feed: status %d, want 404", code)
	}

	// A shared feed cannot be edited by one subscriber.
	if code, _ := doRawAs(t, testUserAlice, http.MethodPut, "/feed/"+sid, map[string]string{"url": "https://example.com/shared2", "title": "S"}); code != http.StatusConflict {
		t.Errorf("editing a shared feed: status %d, want 409", code)
	}

	// Deleting a shared feed only removes it for that user.
	if code, body := doRawAs(t, testUserAlice, http.MethodDelete, "/feed/"+sid+"/purge", nil); code != http.StatusNoContent {
		t.Fatalf("alice deleting shared feed: status %d, body %s", code, body)
	}
	if got := feedIdsOf(t, testUserAlice); got[shared] {
		t.Error("shared feed still in alice's list after she deleted it")
	}
	if got := feedIdsOf(t, testUserBob); !got[shared] {
		t.Error("shared feed disappeared from bob's list")
	}
	if !miniRedis.Exists("news:s1") || list.Count(int(shared)) != 2 {
		t.Error("articles of a still-subscribed feed were purged")
	}
	if miniRedis.Exists("read_mark:" + testUserAlice + ":" + sid) {
		t.Error("alice's read marks for the feed were not cleared")
	}

	// The last subscriber's delete purges the feed and its articles.
	if code, body := doRawAs(t, testUserBob, http.MethodDelete, "/feed/"+sid+"/purge", nil); code != http.StatusNoContent {
		t.Fatalf("bob deleting now-unshared feed: status %d, body %s", code, body)
	}
	if miniRedis.Exists("news:s1") || list.Count(int(shared)) != 0 {
		t.Error("articles of an unsubscribed feed were not purged")
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
