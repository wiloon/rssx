package rss

import (
	"net/http"
	"strconv"

	"rssx/feeds"
	"rssx/utils/jwt"
	log "rssx/utils/logger"

	"github.com/gin-gonic/gin"
)

// SyncAll triggers an immediate sync of the feeds the current user subscribes to.
func SyncAll(c *gin.Context) {
	userId := jwt.UserIdFromContext(c)
	log.Infof("manual sync started for user %s", userId)
	go syncUserFeeds(userId)
	c.JSON(http.StatusOK, gin.H{"message": "sync started"})
}

// SyncOne triggers an immediate sync of a single feed the current user
// subscribes to.
func SyncOne(c *gin.Context) {
	idStr := c.Param("id")
	feedId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feed id"})
		return
	}

	feedList := feeds.FindUserFeeds(jwt.UserIdFromContext(c))
	for _, f := range *feedList {
		if int(f.Id) == feedId {
			log.Infof("manual sync feed triggered, id: %d", feedId)
			go syncOneFeed(f)
			c.JSON(http.StatusOK, gin.H{"message": "sync started", "feed_id": feedId})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "feed not found"})
}
