package feeds

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"rssx/feed"

	"github.com/gin-gonic/gin"
)

// mockFeedRepository is a test double for FeedRepository.
type mockFeedRepository struct {
	findOrCreateFn func(title, url string) (feed.Feed, error)
	isSubscribedFn func(userID string, feedID int64) (bool, error)
	subscribeFn    func(userID string, feedID int64) error
	unsubscribeFn  func(userID string, feedID int64) (bool, error)
	findByUserIDFn func(userID string) ([]feed.Feed, error)
	findByIDFn     func(feedID int64) (feed.Feed, bool, error)
	updateFn       func(feedID int64, title, url string) (feed.Feed, bool, error)
	deleteFn       func(feedID int64) (bool, error)
	subscribersFn  func(feedID int64) ([]string, error)
}

func (m *mockFeedRepository) FindByUserID(userID string) ([]feed.Feed, error) {
	return m.findByUserIDFn(userID)
}
func (m *mockFeedRepository) FindByID(feedID int64) (feed.Feed, bool, error) {
	return m.findByIDFn(feedID)
}
func (m *mockFeedRepository) Update(feedID int64, title, url string) (feed.Feed, bool, error) {
	return m.updateFn(feedID, title, url)
}
func (m *mockFeedRepository) Delete(feedID int64) (bool, error) {
	return m.deleteFn(feedID)
}
func (m *mockFeedRepository) Subscribers(feedID int64) ([]string, error) {
	return m.subscribersFn(feedID)
}
func (m *mockFeedRepository) FindOrCreateByURL(title, url string) (feed.Feed, error) {
	return m.findOrCreateFn(title, url)
}
func (m *mockFeedRepository) IsSubscribed(userID string, feedID int64) (bool, error) {
	return m.isSubscribedFn(userID, feedID)
}
func (m *mockFeedRepository) Subscribe(userID string, feedID int64) error {
	return m.subscribeFn(userID, feedID)
}
func (m *mockFeedRepository) Unsubscribe(userID string, feedID int64) (bool, error) {
	return m.unsubscribeFn(userID, feedID)
}

func newTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/feeds/detail", h.ListFeeds)
	r.POST("/feed", h.AddFeed)
	r.PUT("/feed/:id", h.UpdateFeed)
	r.DELETE("/feed/:id", h.RemoveFeed)
	return r
}

// --- AddFeed tests ---

func TestAddFeed_InvalidJSON(t *testing.T) {
	h := NewHandler(&mockFeedRepository{})
	r := newTestRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/feed", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAddFeed_MissingURL(t *testing.T) {
	h := NewHandler(&mockFeedRepository{})
	r := newTestRouter(h)

	body, _ := json.Marshal(map[string]string{"title": "Test"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/feed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAddFeed_InvalidURLScheme(t *testing.T) {
	h := NewHandler(&mockFeedRepository{})
	r := newTestRouter(h)

	body, _ := json.Marshal(map[string]string{"url": "ftp://bad.com/feed", "title": "Test"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/feed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAddFeed_MissingTitle(t *testing.T) {
	h := NewHandler(&mockFeedRepository{})
	r := newTestRouter(h)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com/feed"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/feed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAddFeed_AlreadySubscribed(t *testing.T) {
	mock := &mockFeedRepository{
		findOrCreateFn: func(title, url string) (feed.Feed, error) {
			return feed.Feed{Id: 1, Title: title, Url: url}, nil
		},
		isSubscribedFn: func(userID string, feedID int64) (bool, error) {
			return true, nil
		},
	}
	h := NewHandler(mock)
	r := newTestRouter(h)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com/feed", "title": "Example"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/feed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestAddFeed_Success(t *testing.T) {
	mock := &mockFeedRepository{
		findOrCreateFn: func(title, url string) (feed.Feed, error) {
			return feed.Feed{Id: 7, Title: title, Url: url}, nil
		},
		isSubscribedFn: func(userID string, feedID int64) (bool, error) {
			return false, nil
		},
		subscribeFn: func(userID string, feedID int64) error {
			return nil
		},
	}
	h := NewHandler(mock)
	r := newTestRouter(h)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com/feed", "title": "Example"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/feed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var resp feed.Feed
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not a feed: %v", err)
	}
	if resp.Id != 7 {
		t.Errorf("expected feed id 7, got %d", resp.Id)
	}
}

// --- RemoveFeed tests ---

func TestRemoveFeed_InvalidID(t *testing.T) {
	h := NewHandler(&mockFeedRepository{})
	r := newTestRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/feed/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRemoveFeed_NotFound(t *testing.T) {
	mock := &mockFeedRepository{
		unsubscribeFn: func(userID string, feedID int64) (bool, error) {
			return false, nil
		},
	}
	h := NewHandler(mock)
	r := newTestRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/feed/42", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRemoveFeed_RepoError(t *testing.T) {
	mock := &mockFeedRepository{
		unsubscribeFn: func(userID string, feedID int64) (bool, error) {
			return false, errors.New("db failure")
		},
	}
	h := NewHandler(mock)
	r := newTestRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/feed/42", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRemoveFeed_Success(t *testing.T) {
	mock := &mockFeedRepository{
		unsubscribeFn: func(userID string, feedID int64) (bool, error) {
			return true, nil
		},
	}
	h := NewHandler(mock)
	r := newTestRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/feed/42", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

// --- ListFeeds tests ---

func TestListFeeds_Success(t *testing.T) {
	mock := &mockFeedRepository{
		findByUserIDFn: func(userID string) ([]feed.Feed, error) {
			return []feed.Feed{
				{Id: 1, Title: "Hacker News", Url: "https://hnrss.org/newest"},
				{Id: 2, Title: "r/golang", Url: "https://www.reddit.com/r/golang/.rss"},
			}, nil
		},
	}
	r := newTestRouter(NewHandler(mock))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/feeds/detail", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []feed.Feed
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not a feed list: %v", err)
	}
	if len(resp) != 2 || resp[0].Url != "https://hnrss.org/newest" {
		t.Errorf("unexpected feed list: %+v", resp)
	}
}

// --- UpdateFeed tests ---

func doUpdate(t *testing.T, h *Handler, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	r := newTestRouter(h)
	var buf *bytes.Buffer
	switch b := body.(type) {
	case string:
		buf = bytes.NewBufferString(b)
	default:
		raw, _ := json.Marshal(b)
		buf = bytes.NewBuffer(raw)
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, path, buf)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestUpdateFeed_InvalidID(t *testing.T) {
	w := doUpdate(t, NewHandler(&mockFeedRepository{}), "/feed/abc",
		map[string]string{"title": "X", "url": "https://example.com/feed"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateFeed_InvalidBody(t *testing.T) {
	w := doUpdate(t, NewHandler(&mockFeedRepository{}), "/feed/1", "not-json")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateFeed_MissingURL(t *testing.T) {
	w := doUpdate(t, NewHandler(&mockFeedRepository{}), "/feed/1",
		map[string]string{"title": "X"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateFeed_NotFound(t *testing.T) {
	mock := &mockFeedRepository{
		updateFn: func(feedID int64, title, url string) (feed.Feed, bool, error) {
			return feed.Feed{}, false, nil
		},
	}
	w := doUpdate(t, NewHandler(mock), "/feed/9",
		map[string]string{"title": "X", "url": "https://example.com/feed"})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateFeed_URLConflict(t *testing.T) {
	mock := &mockFeedRepository{
		updateFn: func(feedID int64, title, url string) (feed.Feed, bool, error) {
			return feed.Feed{}, true, ErrURLConflict
		},
	}
	w := doUpdate(t, NewHandler(mock), "/feed/1",
		map[string]string{"title": "X", "url": "https://example.com/taken"})
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestUpdateFeed_Success(t *testing.T) {
	mock := &mockFeedRepository{
		updateFn: func(feedID int64, title, url string) (feed.Feed, bool, error) {
			return feed.Feed{Id: feedID, Title: title, Url: url}, true, nil
		},
	}
	w := doUpdate(t, NewHandler(mock), "/feed/5",
		map[string]string{"title": "New Title", "url": "https://example.com/feed"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp feed.Feed
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not a feed: %v", err)
	}
	if resp.Id != 5 || resp.Title != "New Title" {
		t.Errorf("unexpected feed: %+v", resp)
	}
}
