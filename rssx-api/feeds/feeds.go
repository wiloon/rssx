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
	"rssx/utils/jwt"
	log "rssx/utils/logger"
)

// FindUserFeeds returns feeds subscribed by one user.
// Kept for backward compatibility with packages that call it directly (e.g. rss/handler.go).
func FindUserFeeds(userId string) *[]feed.Feed {
	feeds := &[]feed.Feed{}
	common.DB.Table("user_feeds").Select("feeds.id,feeds.title,feeds.url").Joins("join feeds on user_feeds.feed_id = feeds.id").Where("user_id = ?", userId).Order("user_feeds.sort desc").Find(feeds)
	return feeds
}

// FindSubscribedFeeds returns every feed that at least one user subscribes to.
// Background sync and GC work on this set; feeds nobody subscribes to are left alone.
func FindSubscribedFeeds() *[]feed.Feed {
	feeds := &[]feed.Feed{}
	common.DB.Table("feeds").Select("feeds.id,feeds.title,feeds.url").
		Where("feeds.id IN (?)", common.DB.Table("user_feeds").Select("feed_id")).
		Order("feeds.id").Find(feeds)
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
	userId := jwt.UserIdFromContext(c)
	feedsList := []feed.Feed{{Id: -1, Title: "All", Url: ""}}
	userFeeds, err := h.repo.FindByUserID(userId)
	if err != nil {
		log.Errorf("failed to load feed list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	feedIds := make([]int, len(userFeeds))
	for i, v := range userFeeds {
		feedIds[i] = int(v.Id)
	}
	unreadCounts := list.FeedUnreadCounts(userId, feedIds)
	var totalUnread int64
	for _, v := range userFeeds {
		unread := unreadCounts[int(v.Id)]
		totalUnread += unread
		v.Title = v.Title + " - " + strconv.Itoa(int(unread))
		feedsList = append(feedsList, v)
	}
	feedsList[0].Title = "All - " + strconv.FormatInt(totalUnread, 10)
	c.JSON(http.StatusOK, feedsList)
}

// ListFeeds returns the current user's feeds with their editable fields
// (id, title, url) and no unread-count decoration — the shape the feed
// management page needs. GET /feeds/detail
func (h *Handler) ListFeeds(c *gin.Context) {
	userFeeds, err := h.repo.FindByUserID(jwt.UserIdFromContext(c))
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

// AddFeed subscribes the current user to an RSS feed. Feeds are shared: if the
// URL is already known, the user is subscribed to the existing feed.
// POST /feed
func (h *Handler) AddFeed(c *gin.Context) {
	userId := jwt.UserIdFromContext(c)
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

	subscribed, err := h.repo.IsSubscribed(userId, f.Id)
	if err != nil {
		log.Errorf("failed to check subscription: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if subscribed {
		c.JSON(http.StatusConflict, gin.H{"error": "already subscribed to this feed"})
		return
	}

	if err := h.repo.Subscribe(userId, f.Id); err != nil {
		if errors.Is(err, ErrAlreadySubscribed) {
			c.JSON(http.StatusConflict, gin.H{"error": "already subscribed to this feed"})
			return
		}
		log.Errorf("failed to create user_feed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusCreated, f)
}

// RemoveFeed unsubscribes the current user from a feed and drops their read
// state for it. The feed and its articles stay for other subscribers.
// DELETE /feed/:id
func (h *Handler) RemoveFeed(c *gin.Context) {
	userId := jwt.UserIdFromContext(c)
	feedId, ok := parseFeedID(c)
	if !ok {
		return
	}

	found, err := h.repo.Unsubscribe(userId, feedId)
	if err != nil {
		log.Errorf("failed to delete user_feed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}
	list.ClearReadState(userId, int(feedId))

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

// subscribersOf loads the feed's subscribers and reports whether userId is one
// of them. On error it writes a 500 response itself and returns ok=false.
func (h *Handler) subscribersOf(c *gin.Context, feedId int64, userId string) (subscribers []string, isSubscriber, ok bool) {
	subscribers, err := h.repo.Subscribers(feedId)
	if err != nil {
		log.Errorf("failed to load subscribers for feed %d: %v", feedId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return nil, false, false
	}
	for _, s := range subscribers {
		if s == userId {
			return subscribers, true, true
		}
	}
	return subscribers, false, true
}

// UpdateFeed changes a feed's title and URL. Feeds are shared, so only a user
// who is the feed's sole subscriber may edit it.
// PUT /feed/:id
func (h *Handler) UpdateFeed(c *gin.Context) {
	userId := jwt.UserIdFromContext(c)
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

	subscribers, isSubscriber, ok := h.subscribersOf(c, feedId, userId)
	if !ok {
		return
	}
	if !isSubscriber {
		c.JSON(http.StatusNotFound, gin.H{"error": "feed not found"})
		return
	}
	if len(subscribers) > 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "this feed is shared with other users and cannot be edited; add the new URL as a separate feed instead"})
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

// PurgeFeed removes a feed from the current user's list. When nobody else
// subscribes to it, the feed row and all of its articles are deleted as well;
// otherwise only this user's subscription and read state go.
// DELETE /feed/:id/purge
func (h *Handler) PurgeFeed(c *gin.Context) {
	userId := jwt.UserIdFromContext(c)
	feedId, ok := parseFeedID(c)
	if !ok {
		return
	}

	subscribers, isSubscriber, ok := h.subscribersOf(c, feedId, userId)
	if !ok {
		return
	}
	if !isSubscriber {
		c.JSON(http.StatusNotFound, gin.H{"error": "feed not found"})
		return
	}

	if len(subscribers) > 1 {
		if _, err := h.repo.Unsubscribe(userId, feedId); err != nil {
			log.Errorf("failed to unsubscribe user %s from feed %d: %v", userId, feedId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		list.ClearReadState(userId, int(feedId))
		c.Status(http.StatusNoContent)
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
