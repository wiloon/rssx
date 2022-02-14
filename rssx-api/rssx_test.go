package main

import (
	"rssx/data"
	"rssx/rss"
	log "rssx/utils/logger"
	"testing"
)

func Test0(t *testing.T) {
	log.Init()
	feeds := data.FindFeeds()
	for _, feed := range feeds {
		rss.SyncFeed(feed)
	}
}
