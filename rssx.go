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
	router.GET("/news-previous", PreviousNews)
	router.GET("/mark-read", MarkReadNews)

	err := router.Run()
	handleErr(err)

	utils.WaitSignals()
}

func handleErr(e error) {
	if e != nil {
		log.Info(e.Error())
	}
}

func MarkReadNews(c *gin.Context) {

	feedId, _ := strconv.Atoi(c.Query("feedId"))
	readIndex := list.GetLatestReadIndex(user.DefaultId, feedId)
	// reset read index
	newIndex := readIndex + list.PageSize //新已读=旧值加每页数量
	count := list.Count(feedId)
	if newIndex > count {
		newIndex = count - 1
	}
	log.Infof("mark page as read, feed id: %v,  last read index: %v, new index: %v", feedId, readIndex, newIndex)

	list.SetReadIndex(0, feedId, newIndex) //save
	// del read mark set,按feed删除
	news.DelReadMark(0, feedId)

	// load next page
	newsList := list.LoadNewsListByFeed(feedId)
	c.JSON(200, newsList)
}
func PreviousNews(c *gin.Context) {
	currentNewsId := c.Query("currentId")
	feedId, _ := strconv.Atoi(c.Query("feedId"))
	log.Debugf(" load previous news feed id:%v, news id:%v", feedId, currentNewsId)
	index := list.FindIndexById(feedId, currentNewsId)
	newsIds := list.FindNewsListByRange(list.NewsListKey(feedId), index-1, index-1)
	previousNewsId := newsIds[0]
	n := news.New(previousNewsId)
	n.FeedId = int64(feedId)
	n.Load()
	nextNewsId := list.FindNextId(feedId, previousNewsId)
	n.NextId = nextNewsId
	c.JSON(200, n)

}
func LoadNews(c *gin.Context) {

	newsId := c.Query("id")
	feedId, _ := strconv.Atoi(c.Query("feedId"))
	log.Debugf(" load news feed id:%v, news id:%v", feedId, newsId)

	n := news.New(newsId)
	n.FeedId = int64(feedId)
	n.Load()
	log.Info("news:" + n.Title)

	nextNewsId := list.FindNextId(feedId, newsId)
	n.NextId = nextNewsId

	log.Info("show news:", n.Title, ", next:", n.NextId)
	n.MarkRead(0)
	c.JSON(200, n)

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
