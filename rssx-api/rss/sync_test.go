package rss

import (
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/golang/mock/gomock"
	"rssx/feed"
	"rssx/feeds/mocks"
	"rssx/utils/config"
	"rssx/utils/logger"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	config.LoadConfigByPath("../config.toml")
	logger.Init("console,file", "debug", "rssx")
	mockCtl := gomock.NewController(t)

	mockRssFeeds := mocks.NewMockRssFeeds(mockCtl)

	testFeeds := make([]*feed.Feed, 2)
	testFeeds[0] = &feed.Feed{Id: 12, Title: "os china", Url: "https://www.oschina.net/news/rss"}
	testFeeds[1] = &feed.Feed{Id: 13, Title: "draveness", Url: "https://draveness.me/feed.xml"}
	mockRssFeeds.EXPECT().GetAllFeedList().Return(testFeeds)

	// redis
	s := miniredis.RunT(t)

	// Optionally set some keys your code expects:
	s.Set("foo", "bar")
	s.HSet("some", "other", "key")
	s.Addr()

	fmt.Printf("mini redis addr: %s\n", s.Addr())

	syncFeeds(mockRssFeeds)

	time.Sleep(10 * time.Second)
}

func TestTime(t *testing.T) {
	fmt.Println(time.Now().UnixNano())
	fmt.Println(time.Now().UnixNano())
}
