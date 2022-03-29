package main

import (
	"github.com/gin-gonic/gin"
	"rssx/feed/news/list"
	"rssx/feeds"
	"rssx/rss"
	"rssx/user"
	"rssx/utils"
	"rssx/utils/config"
	log "rssx/utils/logger"
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

	router.GET("/feeds", feeds.LoadFeedList)
	router.GET("/news-list", list.LoadNewsList)
	router.GET("/news", list.LoadArticles)
	router.GET("/previous-news", list.PreviousArticle)
	router.GET("/mark-read", list.MarkWholePageAsRead)
	router.POST("/login", user.Login)
	router.POST("/register", user.Register)

	err := router.Run(":8080")
	if err != nil {
		log.Errorf("failed to start rssx: %v", err)
	}
	log.Info("rssx started and listening default port of gin")
	utils.WaitSignals()
}
