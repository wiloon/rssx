package rss

import (
	"fmt"
	"rssx/utils/config"
	"rssx/utils/logger"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	config.LoadConfigByPath("../config.toml")
	logger.Init("console,file", "debug", "rssx")

	syncFeeds()

	time.Sleep(10 * time.Second)
}

func TestTime(t *testing.T) {
	fmt.Println(time.Now().UnixNano())
	fmt.Println(time.Now().UnixNano())
}
