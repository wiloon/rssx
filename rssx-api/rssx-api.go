package main

import (
	"os"
	"rssx/common"
	"rssx/feed/news/list"
	"rssx/feeds"
	"rssx/legacy"
	"rssx/rss"
	"rssx/user"
	"rssx/utils"
	"rssx/utils/config"
	"rssx/utils/jwt"
	log "rssx/utils/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Init("CONSOLE", config.GetString("log.level", "info"), "rssx-api")

	if config.GetString("rssx.security-key", "") == "" {
		log.Error("rssx.security-key (env RSSX_SECURITY_KEY) is empty; refusing to start")
		os.Exit(1)
	}

	if mode := config.GetString("gin.mode", gin.ReleaseMode); mode != "" {
		gin.SetMode(mode)
	}

	if err := legacy.MigrateLegacyUserData(common.DB, config.GetString("rssx.legacy-owner", "")); err != nil {
		log.Errorf("failed to migrate legacy single-user data: %v", err)
	}

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

	router.POST("/login", user.Login)
	router.POST("/register", user.Register)

	authed := router.Group("/", jwt.RequireAuth())
	feedHandler := feeds.NewHandler(feeds.NewGormFeedRepository(common.DB))
	authed.GET("/feeds", feedHandler.LoadFeedList)
	authed.GET("/feeds/detail", feedHandler.ListFeeds)
	authed.POST("/feed", feedHandler.AddFeed)
	authed.PUT("/feed/:id", feedHandler.UpdateFeed)
	authed.DELETE("/feed/:id", feedHandler.RemoveFeed)
	authed.DELETE("/feed/:id/purge", feedHandler.PurgeFeed)
	authed.POST("/sync", rss.SyncAll)
	authed.POST("/sync/:id", rss.SyncOne)
	authed.GET("/news-list", list.LoadNewsList)
	authed.GET("/news", list.LoadArticles)
	authed.GET("/previous-news", list.PreviousArticle)
	authed.GET("/mark-read", list.MarkWholePageAsRead)
	return router
}
