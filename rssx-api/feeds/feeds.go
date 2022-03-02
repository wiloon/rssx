package feeds

import (
	"rssx/data"
	"rssx/feed"
)

type RssFeeds interface {
	GetAllFeedList() []*feed.Feed
}

type DefaultFeeds struct {
}

// GetAllFeedList get all feeds
func (f *DefaultFeeds) GetAllFeedList() []*feed.Feed {
	stmt := "select feed_id,title,url from feed where deleted=?"
	result := data.Rssx().Find(stmt, []interface{}{0}...)
	var feeds []*feed.Feed
	for _, v := range result {
		feeds = append(feeds, &feed.Feed{
			Id:    v["feed_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url:   string(v["url"].([]uint8)),
		})
	}
	return feeds
}
