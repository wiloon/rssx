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

func MarkWholePageAsRead(c *gin.Context) {

	feedId, _ := strconv.Atoi(c.Query("feedId"))
	readIndex := list.GetLatestReadIndex(user.DefaultId, feedId)
	// reset read index
	newIndex := readIndex + list.PageSize //新已读=旧值加每页数量
	count := list.Count(feedId)
	if newIndex > count {
		newIndex = count - 1
	}
	log.Infof("mark page as read, feed id: %v,  last read index: %v, new index: %v, list count: %v",
		feedId, readIndex, newIndex, count)

	list.SetReadIndex(0, feedId, newIndex) //save
	// del read mark set,按feed删除
	news.DelReadMark(0, feedId)

	// load next page
	newsList := list.LoadNewsListByFeed(feedId)
	c.JSON(200, newsList)
}

func PreviousNews(c *gin.Context) {
	currentNewsId := c.Query("newsId")
	feedId, _ := strconv.Atoi(c.Query("feedId"))
	log.Debugf(" load previous news feed id:%v, news id:%v", feedId, currentNewsId)
	index := list.FindIndexById(feedId, currentNewsId)
	newsIds := list.FindNewsListByRange(list.NewsListKey(feedId), index-1, index-1)
	previousNewsId := newsIds[0]
	previousNews := news.New(previousNewsId)
	previousNews.FeedId = int64(feedId)
	previousNews.Load()
	nextNewsId := list.FindNextId(feedId, previousNewsId)
	previousNews.NextId = nextNewsId
	c.JSON(200, previousNews)

}

/*
	LoadNews: load one news
    按id加载一条新闻
*/
func LoadNews(c *gin.Context) {
	feedId, _ := strconv.Atoi(c.Query("feedId"))
	newsId := c.Query("id")

	n := news.New(newsId)
	n.FeedId = int64(feedId)
	n.Load()
	log.Debugf("load one news, feed id:%v, news id:%v, title: %s", feedId, newsId, n.Title)

	nextNewsId := list.FindNextId(feedId, newsId)
	n.NextId = nextNewsId

	log.Info("show news:", n.Title, ", next id:", n.NextId)

	// 加载新的一条新闻时要维护已读未读的边界 和 不连续的已读记录
	// 用户当前已读索引
	currentUserReadIndex := list.GetLatestReadIndex(user.DefaultId, feedId)
	// 当前新闻的索引
	currentNewsIndex := list.FindIndexById(feedId, newsId)
	n.MarkRead(0)
	log.Debugf("currentUserReadIndex: %v, currentNewsIndex: %v", currentUserReadIndex, currentNewsIndex)

	nextUnReadIndex := findNextUserUnReadIndex(feedId, currentUserReadIndex)
	log.Debugf("currentUserReadIndex: %v, nextUnReadIndex: %v", currentUserReadIndex, nextUnReadIndex)
	if currentUserReadIndex == nextUnReadIndex {
		// 已读位置不连续，记录到已读集合
		n.MarkRead(0)
	} else {
		//已读新闻是连续的，直接维护已读位置边界
		//更新用户已读索引
		list.SetReadIndex(0, feedId, nextUnReadIndex)
	}
	c.JSON(200, n)

}

/**
找到用户下一个未读索引
*/
func findNextUserUnReadIndex(feedId int, currentNewsIndex int64) int64 {
	log.Debugf("findNextUserUnReadIndex, feed id: %v, index: %v", feedId, currentNewsIndex)
	var result int64
	nextNewsIndex := currentNewsIndex + 1
	nextNewsId := list.FinOneNewsByIndex(nextNewsIndex, feedId)

	if nextNewsId == "" {
		result = currentNewsIndex
	} else {
		nextNews := news.New(nextNewsId)
		nextNews.FeedId = int64(feedId)
		if nextNews.IsRead(user.DefaultId) {
			result = findNextUserUnReadIndex(feedId, nextNewsIndex)
		} else {
			// 找到一条未读新闻，退出
			result = currentNewsIndex
		}
	}

	log.Debugf("findNextUserUnReadIndex, feed id: %v, index: %v, result: %v", feedId, currentNewsIndex, result)
	return result
}
func checkIfAllPreviousNewsIsRead(feedId int, newsId string) bool {
	result := false
	// 用户当前已读索引
	currentUserReadIndex := list.GetLatestReadIndex(user.DefaultId, feedId)
	// 当前新闻的索引
	currentNewsIndex := list.FindIndexById(feedId, newsId)
	log.Debugf("checkIfAllPreviousNewsIsRead, currentUserReadIndex: %v, currentNewsIndex: %v", currentUserReadIndex, currentNewsIndex)
	previousNewsIndex := currentNewsIndex - 1
	if previousNewsIndex == currentUserReadIndex {
		// 一直往前找，直到用户未读索引都是未读新闻，退出
		result = true
	} else {
		// 检查上一条新闻是不是已读
		previousNewsId := list.FinOneNewsByIndex(previousNewsIndex, feedId)
		previousNews := news.New(previousNewsId)
		previousNews.FeedId = int64(feedId)
		if previousNews.IsRead(user.DefaultId) {
			result = checkIfAllPreviousNewsIsRead(feedId, previousNewsId)
		} else {
			// 找到一条未读新闻，退出
			result = false
		}
	}
	log.Debugf("checkIfAllPreviousNewsIsRead, currentUserReadIndex: %v, currentNewsIndex: %v, result: %v", currentUserReadIndex, currentNewsIndex, result)
	return result
}
func LoadNewsList(c *gin.Context) {
	feedIdStr := c.Query("id")
	feedId, _ := strconv.Atoi(feedIdStr)
	log.Debugf("load news list by feed id: %v", feedId)

	newsList := list.LoadNewsListByFeed(feedId)

	c.JSON(200, newsList)

}

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
