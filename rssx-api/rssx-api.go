package main

import (
	"github.com/gin-gonic/gin"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/feeds"
	"rssx/rss"
	"rssx/user"
	"rssx/utils"
	"rssx/utils/config"
	log "rssx/utils/logger"
	"strconv"
)

func main() {
	log.Init("file", "debug", "rssx-api")

	//定时同步文章列表， rss源>redis
	syncAuto := config.GetBoolWithDefaultValue("rssx.rss-sync-auto", false)
	log.Infof("sync auto: %t", syncAuto)
	if syncAuto {
		go rss.Sync()
	}

	//定时清理缓存
	go rss.Gc()

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/feeds", LoadFeedList)
	router.GET("/news-list", list.LoadNewsList)
	router.GET("/news", list.LoadArticles)
	router.GET("/previous-news", list.PreviousArticle)
	router.GET("/mark-read", list.MarkWholePageAsRead)
	router.POST("/login", user.Login)
	router.POST("/register", user.Register)

	err := router.Run(":8080")
	handleErr(err)
	log.Info("rssx started and listening default port of gin")
	utils.WaitSignals()
}

func handleErr(e error) {
	if e != nil {
		log.Info(e.Error())
	}
}

func LoadFeedList(c *gin.Context) {
	log.Debug("load user feed list")
	feedsList := []feed.Feed{{Id: -1, Title: "All", Url: ""}}
	tmp := feeds.FindUserFeeds(user.DefaultId)
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
