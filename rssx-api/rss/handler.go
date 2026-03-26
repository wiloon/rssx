package rss

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"rssx/feeds"
	log "rssx/utils/logger"
)

// SyncAll triggers an immediate sync of all feeds.
func SyncAll(c *gin.Context) {
	log.Info("manual sync all feeds triggered")
	go syncFeeds()
	c.JSON(http.StatusOK, gin.H{"message": "sync started"})
}

// SyncOne triggers an immediate sync of a single feed by ID.
func SyncOne(c *gin.Context) {
	idStr := c.Param("id")
	feedId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feed id"})
		return
	}

	feedList := feeds.FindUserFeeds("0")
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
