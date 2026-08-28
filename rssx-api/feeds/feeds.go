package feeds

import (
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

// addFeedRequest is the request body for POST /feed.
type addFeedRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// AddFeed subscribes the default user to a new RSS feed.
// POST /feed
func (h *Handler) AddFeed(c *gin.Context) {
	var req addFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)

	if req.URL == "" || (!strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required and must start with http:// or https://"})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
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
