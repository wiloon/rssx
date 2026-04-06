package main

import (
	"os"
	"rssx/common"
	"rssx/feed/news/list"
	"rssx/feeds"
	"rssx/rss"
	"rssx/user"
	"rssx/utils"
	"rssx/utils/config"
	log "rssx/utils/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Init("CONSOLE", "debug", "rssx-api")

	//定时同步文章列表， rss源>redis
	syncAuto := config.GetBoolWithDefaultValue("rssx.rss-sync-auto", false)
	log.Infof("sync auto: %t", syncAuto)
	if syncAuto {
		go rss.Sync()
	}

	//定时清理缓存
	go rss.Gc()

	router := setupRouter()
	err := router.Run(":8080")
	if err != nil {
		log.Errorf("failed to start rssx: %v", err)
		os.Exit(1)
	}
	log.Info("rssx started and listening default port of gin")
	utils.WaitSignals()
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	feedHandler := feeds.NewHandler(feeds.NewGormFeedRepository(common.DB))
	router.GET("/feeds", feedHandler.LoadFeedList)
	router.POST("/feed", feedHandler.AddFeed)
	router.DELETE("/feed/:id", feedHandler.RemoveFeed)
    router.POST("/sync", rss.SyncAll)
    router.POST("/sync/:id", rss.SyncOne)
	router.GET("/news-list", list.LoadNewsList)
	router.GET("/news", list.LoadArticles)
	router.GET("/previous-news", list.PreviousArticle)
	router.GET("/mark-read", list.MarkWholePageAsRead)
	router.POST("/login", user.Login)
	router.POST("/register", user.Register)
	return router
}
