package main

import (
	log "github.com/wiloon/pingd-log/logconfig/zaplog"
	"rssx/data"
	"rssx/rss"
	"testing"
)

func Test0(t *testing.T) {
	log.Init()
	feeds := data.FindFeeds()
	for _, feed := range feeds {
		rss.SyncFeed(feed)
	}
}
