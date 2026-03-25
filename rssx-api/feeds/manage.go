package feeds

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"rssx/common"
	"rssx/feed"
	log "rssx/utils/logger"
)

type addFeedRequest struct {
	Title string `json:"title" binding:"required"`
	Url   string `json:"url" binding:"required"`
}

// AddFeed adds a new RSS feed and subscribes the default user to it.
// If the feed URL already exists, it reuses the existing feed record.
func AddFeed(c *gin.Context) {
	var req addFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "title and url are required"})
		return
	}

	// Check if feed URL already exists
	existing := &common.Feed{}
	result := common.DB.Where("url = ?", req.Url).First(existing)
	var feedId int64
	if result.Error == nil {
		// Feed already exists, reuse it
		feedId = existing.Id
		log.Infof("feed already exists, reusing id=%d url=%s", feedId, req.Url)
	} else {
		// Create new feed
		newFeed := &common.Feed{Title: req.Title, Url: req.Url}
		if err := common.DB.Create(newFeed).Error; err != nil {
			log.Errorf("failed to create feed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to create feed"})
			return
		}
		feedId = newFeed.Id
		log.Infof("created new feed id=%d url=%s", feedId, req.Url)
	}

	// Check if user already subscribed
	userFeed := &common.UserFeed{}
	subResult := common.DB.Where("user_id = ? AND feed_id = ?", "0", feedId).First(userFeed)
	if subResult.Error == nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "already subscribed", "id": feedId})
		return
	}

	// Subscribe default user
	sub := &common.UserFeed{UserId: "0", FeedId: feedId}
	if err := common.DB.Create(sub).Error; err != nil {
		log.Errorf("failed to subscribe feed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to subscribe"})
		return
	}

	c.JSON(http.StatusOK, feed.Feed{Id: feedId, Title: req.Title, Url: req.Url})
}

// DeleteFeed removes the default user's subscription to the specified feed.
func DeleteFeed(c *gin.Context) {
	idStr := c.Param("id")
	feedId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid feed id"})
		return
	}

	result := common.DB.Where("user_id = ? AND feed_id = ?", "0", feedId).Delete(&common.UserFeed{})
	if result.Error != nil {
		log.Errorf("failed to delete user feed subscription: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to delete feed"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "feed subscription not found"})
		return
	}

	log.Infof("deleted feed subscription feed_id=%d", feedId)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
}
