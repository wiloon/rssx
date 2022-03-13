package feeds

import (
	"rssx/common"
	"rssx/feed"
	zapLogger "rssx/utils/logger"
)

func FindUserFeeds(userId string) *[]feed.Feed {
	feeds := &[]feed.Feed{}
	common.DB.Table("user_feeds").Select("feeds.id,feeds.title,feeds.url").Joins("join feeds on user_feeds.feed_id = feeds.id").Where("user_id = ?", userId).Order("user_feeds.sort desc").Find(feeds)
	return feeds
}

func checkErr(err error) {
	if err != nil {
		zapLogger.Errorf("err: %v", err)
	}
}
