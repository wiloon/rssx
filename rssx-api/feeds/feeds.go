package feeds

import (
	"github.com/gin-gonic/gin"
	"rssx/common"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/user"
	log "rssx/utils/logger"
	"strconv"
)

// FindUserFeeds feeds subscribed by one user
func FindUserFeeds(userId string) *[]feed.Feed {
	feeds := &[]feed.Feed{}
	common.DB.Table("user_feeds").Select("feeds.id,feeds.title,feeds.url").Joins("join feeds on user_feeds.feed_id = feeds.id").Where("user_id = ?", userId).Order("user_feeds.sort desc").Find(feeds)
	return feeds
}
func LoadFeedList(c *gin.Context) {
	log.Debug("load user feed list")
	feedsList := []feed.Feed{{Id: -1, Title: "All", Url: ""}}
	tmp := FindUserFeeds(user.DefaultId)
	log.Info("user feeds: %+v", tmp)
	for _, v := range *tmp {
		log.Debugf("feed: %+v", v)
		count := list.Count(int(v.Id))
		index := list.GetLatestReadIndex(user.DefaultId, int(v.Id))
		unread := count - index - 1
		if unread < 0 {
			unread = 0
		}
		v.Title = v.Title + " - " + strconv.Itoa(int(unread))
		log.Debugf("feed list item: %v", v)
		feedsList = append(feedsList, v)
	}
	c.JSON(200, feedsList)
}
