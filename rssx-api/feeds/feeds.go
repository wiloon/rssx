package feeds

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"rssx/common"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/user"
	log "rssx/utils/logger"
)

// FindUserFeeds returns feeds subscribed by one user.
// Kept for backward compatibility with packages that call it directly (e.g. rss/gc.go).
func FindUserFeeds(userId string) *[]feed.Feed {
	feeds := &[]feed.Feed{}
	common.DB.Table("user_feeds").Select("feeds.id,feeds.title,feeds.url").Joins("join feeds on user_feeds.feed_id = feeds.id").Where("user_id = ?", userId).Order("user_feeds.sort desc").Find(feeds)
	return feeds
}

// Handler holds dependencies for feed HTTP handlers.
type Handler struct {
	repo FeedRepository
}

// NewHandler creates a Handler with the given FeedRepository.
func NewHandler(repo FeedRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) LoadFeedList(c *gin.Context) {
	log.Debug("load user feed list")
	feedsList := []feed.Feed{{Id: -1, Title: "All", Url: ""}}
	userFeeds, err := h.repo.FindByUserID(user.DefaultId)
	if err != nil {
		log.Errorf("failed to load feed list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	feedIds := make([]int, len(userFeeds))
	for i, v := range userFeeds {
		feedIds[i] = int(v.Id)
	}
	unreadCounts := list.FeedUnreadCounts(user.DefaultId, feedIds)

	for _, v := range userFeeds {
		unread := unreadCounts[int(v.Id)]
		v.Title = v.Title + " - " + strconv.Itoa(int(unread))
		feedsList = append(feedsList, v)
	}
	c.JSON(http.StatusOK, feedsList)
}

// ListFeeds returns the default user's feeds with their editable fields
// (id, title, url) and no unread-count decoration — the shape the feed
// management page needs. GET /feeds/detail
func (h *Handler) ListFeeds(c *gin.Context) {
	userFeeds, err := h.repo.FindByUserID(user.DefaultId)
	if err != nil {
		log.Errorf("failed to load feed detail list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusOK, userFeeds)
}

// feedRequest is the request body for POST /feed and PUT /feed/:id.
type feedRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// normalize trims the request fields and validates them the same way for add
// and update. It returns an error message suitable for a 400 response.
func (r *feedRequest) normalize() string {
	r.URL = strings.TrimSpace(r.URL)
	r.Title = strings.TrimSpace(r.Title)
	if r.URL == "" || (!strings.HasPrefix(r.URL, "http://") && !strings.HasPrefix(r.URL, "https://")) {
		return "url is required and must start with http:// or https://"
	}
	if r.Title == "" {
		return "title is required"
	}
	return ""
}

// AddFeed subscribes the default user to a new RSS feed.
// POST /feed
func (h *Handler) AddFeed(c *gin.Context) {
	var req feedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if msg := req.normalize(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	f, err := h.repo.FindOrCreateByURL(req.Title, req.URL)
	if err != nil {
		log.Errorf("failed to upsert feed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	subscribed, err := h.repo.IsSubscribed(user.DefaultId, f.Id)
	if err != nil {
		log.Errorf("failed to check subscription: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if subscribed {
		c.JSON(http.StatusConflict, gin.H{"error": "already subscribed to this feed"})
		return
	}

	if err := h.repo.Subscribe(user.DefaultId, f.Id); err != nil {
		log.Errorf("failed to create user_feed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusCreated, f)
}

// RemoveFeed unsubscribes the default user from a feed.
// DELETE /feed/:id
func (h *Handler) RemoveFeed(c *gin.Context) {
	idStr := c.Param("id")
	feedId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
		return
	}

	found, err := h.repo.Unsubscribe(user.DefaultId, feedId)
	if err != nil {
		log.Errorf("failed to delete user_feed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// parseFeedID reads the :id path param. It writes the 400 response itself and
// returns ok=false when the value is not an integer.
func parseFeedID(c *gin.Context) (int64, bool) {
	feedId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
		return 0, false
	}
	return feedId, true
}

// UpdateFeed changes a feed's title and URL.
// PUT /feed/:id
func (h *Handler) UpdateFeed(c *gin.Context) {
	feedId, ok := parseFeedID(c)
	if !ok {
		return
	}

	var req feedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if msg := req.normalize(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	f, found, err := h.repo.Update(feedId, req.Title, req.URL)
	if err != nil {
		if errors.Is(err, ErrURLConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "another feed already uses this url"})
			return
		}
		log.Errorf("failed to update feed %d: %v", feedId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "feed not found"})
		return
	}

	c.JSON(http.StatusOK, f)
}

// PurgeFeed deletes a feed outright: its row, every subscription to it, and all
// of its articles and read state in Redis.
// DELETE /feed/:id/purge
func (h *Handler) PurgeFeed(c *gin.Context) {
	feedId, ok := parseFeedID(c)
	if !ok {
		return
	}

	subscribers, err := h.repo.Subscribers(feedId)
	if err != nil {
		log.Errorf("failed to load subscribers for feed %d: %v", feedId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	found, err := h.repo.Delete(feedId)
	if err != nil {
		log.Errorf("failed to delete feed %d: %v", feedId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "feed not found"})
		return
	}

	list.PurgeFeed(int(feedId), subscribers)

	c.Status(http.StatusNoContent)
}
