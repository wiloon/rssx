package main

import (
	"github.com/gin-gonic/gin"
	config "github.com/wiloon/pingd-config"
	log "github.com/wiloon/pingd-log/logconfig/zaplog"
	"github.com/wiloon/pingd-utils/utils"
	"rssx/data"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/news"
	"rssx/rss"
	"rssx/user"
	"strconv"
)

func main() {
	log.Init()

	//同步新闻列表， rss源>redis
	if !config.GetBool("rssx.dev-mode") {
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
	router.GET("/news-list", LoadNewsList)
	router.GET("/news", LoadNews)
	router.GET("/previous-news", PreviousNews)
	router.GET("/mark-read", MarkWholePageAsRead)

	err := router.Run()
	handleErr(err)
	log.Info("rssx started and listening default port of gin")
	utils.WaitSignals()
}

func handleErr(e error) {
	if e != nil {
		log.Info(e.Error())
	}
}

<<<<<<< HEAD
// MarkWholePageAsRead 标记整页新闻为已读。
=======
// MarkWholePageAsRead 标记整页新闻为已读
>>>>>>> dc3c267651c63e597f53559c323a5696f4813dc0
func MarkWholePageAsRead(c *gin.Context) {
	feedID, _ := strconv.Atoi(c.Query("feedId"))
	readIndex := list.GetLatestReadIndex(user.DefaultId, feedID)
	// reset read index
	newIndex := readIndex + list.PageSize //新已读=旧值加每页数量
	count := list.Count(feedID)
	if newIndex > count {
		newIndex = count - 1
	}
	log.Infof("mark page as read, feed id: %v,  last read index: %v, new index: %v", feedID, readIndex, newIndex)

	list.SetReadIndex(0, feedID, newIndex) //save
	// del read mark set,按feed删除
	news.DelReadMark(0, feedID)

	// load next page
	newsList := list.LoadNewsListByFeed(feedID)
	c.JSON(200, newsList)
}

// PreviousNews 上一条新闻 
func PreviousNews(c *gin.Context) {
	currentNewsID := c.Query("newsId")
	feedID, _ := strconv.Atoi(c.Query("feedId"))
	log.Debugf(" load previous news feed id:%v, news id:%v", feedID, currentNewsID)
	index := list.FindIndexById(feedID, currentNewsID)
	newsIds := list.FindNewsListByRange(list.NewsListKey(feedID), index-1, index-1)
	previousNewsID := newsIds[0]
	previousNews := news.New(previousNewsID)
	previousNews.FeedId = int64(feedID)
	previousNews.Load()
	nextNewsID := list.FindNextId(feedID, previousNewsID)
	previousNews.NextId = nextNewsID
	c.JSON(200, previousNews)

}

// LoadNews  load one news
func LoadNews(c *gin.Context) {
	feedID, _ := strconv.Atoi(c.Query("feedId"))
	newsID := c.Query("id")

	n := news.New(newsID)
	n.FeedId = int64(feedID)
	n.Load()
	log.Debugf(" load one news, feed id:%v, news id:%v, title: %s", feedID, newsID, n.Title)

	nextNewsID := list.FindNextId(feedID, newsID)
	n.NextId = nextNewsID

	log.Info("show news:", n.Title, ", next:", n.NextId)

	// 加载新的一条新闻时要维护已读未读的边界 和 不连续的已读记录
	// 用户当前已读索引
	currentUserReadIndex := list.GetLatestReadIndex(user.DefaultId, feedID)
	// 当前新闻的索引
	currentNewsIndex := list.FindIndexById(feedID, newsID)
	n.MarkRead(0)
	log.Debugf("currentUserReadIndex: %v, currentNewsIndex: %v", currentUserReadIndex, currentNewsIndex)

	latestReadIndex := findNewestReadIndex(feedID, currentUserReadIndex)
	log.Debugf("currentUserReadIndex: %v, latestReadIndex: %v", currentUserReadIndex, latestReadIndex)
	if currentUserReadIndex == latestReadIndex {
		// 已读位置不连续，记录到已读集合
		n.MarkRead(0)
	} else {
		//已读新闻是连续的，直接维护已读位置边界
		//更新用户已读索引
		list.SetReadIndex(0, feedID, latestReadIndex)
	}
	c.JSON(200, n)

}
func findNewestReadIndex(feedID int, newsIndex int64) int64 {
	log.Debugf("findNewestReadIndex, feed id: %v, index: %v", feedID, newsIndex)
	var result int64
	// 当前新闻的索引
	currentNewsIndex := newsIndex
	nextNewsIndex := currentNewsIndex + 1
	nextNewsID := list.FinOneNewsByIndex(nextNewsIndex, feedID)
	nextNews := news.New(nextNewsID)
	nextNews.FeedId = int64(feedID)
	if nextNews.IsRead(user.DefaultId) {
		result = findNewestReadIndex(feedID, nextNewsIndex)
	} else {
		// 找到一条未读新闻，退出
		result = currentNewsIndex
	}
	log.Debugf("findNewestReadIndex, feed id: %v, index: %v, result: %v", feedID, newsIndex, result)
	return result
}
func checkIfAllPreviousNewsIsRead(feedID int, newsID string) bool {
	result := false
	// 用户当前已读索引
	currentUserReadIndex := list.GetLatestReadIndex(user.DefaultId, feedID)
	// 当前新闻的索引
	currentNewsIndex := list.FindIndexById(feedID, newsID)
	log.Debugf("checkIfAllPreviousNewsIsRead, currentUserReadIndex: %v, currentNewsIndex: %v", currentUserReadIndex, currentNewsIndex)
	previousNewsIndex := currentNewsIndex - 1
	if previousNewsIndex == currentUserReadIndex {
		// 一直往前找，直到用户未读索引都是未读新闻，退出
		result = true
	} else {
		// 检查上一条新闻是不是已读
		previousNewsID := list.FinOneNewsByIndex(previousNewsIndex, feedID)
		previousNews := news.New(previousNewsID)
		previousNews.FeedId = int64(feedID)
		if previousNews.IsRead(user.DefaultId) {
			result = checkIfAllPreviousNewsIsRead(feedID, previousNewsID)
		} else {
			// 找到一条未读新闻，退出
			result = false
		}
	}
	log.Debugf("checkIfAllPreviousNewsIsRead, currentUserReadIndex: %v, currentNewsIndex: %v, result: %v", currentUserReadIndex, currentNewsIndex, result)
	return result
}

// LoadNewsList 加载新闻列表
func LoadNewsList(c *gin.Context) {
	feedIDStr := c.Query("id")
	feedID, _ := strconv.Atoi(feedIDStr)
	log.Debugf("load news list by feed id: %v", feedID)

	newsList := list.LoadNewsListByFeed(feedID)

	c.JSON(200, newsList)

}

// LoadFeedList feed 列表
func LoadFeedList(c *gin.Context) {
	log.Debug("load user feed list")
	feeds := []feed.Feed{{Id: -1, Title: "All", Url: ""}}
	tmp := data.FindUserFeeds(user.DefaultId)

	for _, v := range tmp {
		count := list.Count(int(v.Id))
		index := list.GetLatestReadIndex(0, int(v.Id))
		unread := count - index - 1
		if unread < 0 {
			unread = 0
		}
		v.Title = v.Title + " - " + strconv.Itoa(int(unread))
		log.Debugf("feed list item: %v", v)
		feeds = append(feeds, v)
	}
	c.JSON(200, feeds)
}
