package main

import (
	"encoding/json"
	config "github.com/wiloon/pingd-config"
	"os"

	log "github.com/sirupsen/logrus"
	"net/http"
	"rssx/data"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/news"
	"strconv"
)

const userId = 0

type HttpServer struct {
}

// user feed list
func (server HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("load user feed list")
	feeds := []feed.Feed{{Id: -1, Title: "All", Url: ""}}
	tmp := data.FindUserFeeds(userId)

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

	jsonStr, _ := json.Marshal(feeds)
	log.Info("api feeds:", jsonStr)
	_, _ = w.Write([]byte(jsonStr))
}

type NewsListServer struct {
}

// load news list by feed
func (server NewsListServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	feedId, _ := strconv.Atoi(r.Form.Get("id"))
	log.Debugf("load news list by feed id: %v", feedId)

	newsList := loadNewsListByFeed(feedId)

	jsonStr, _ := json.Marshal(newsList)

	w.Write([]byte(jsonStr))
}

func loadNewsListByFeed(feedId int) []news.News {
	var newsList []news.News
	if feedId == -1 {
		// find all news for all user feeds
		newsList = data.FindAllNewsForUser(userId)
	} else {
		// by feed id
		newsIds := list.FindNewsListByUserFeed(userId, feedId)
		for _, v := range newsIds {
			n := news.New(v)
			n.FeedId = int64(feedId)
			n.LoadTitle()
			n.LoadReadFlag(0)
			// calculate unread count

			newsList = append(newsList, *n)
			log.Debugf("append article: %v", n.Title)
		}
	}
	log.Debugf("new list size: %v", len(newsList))
	return newsList
}

type NewsServer struct {
}

// load news
func (server NewsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	newsId := r.Form.Get("id")
	feedId, _ := strconv.Atoi(r.Form.Get("feedId"))
	log.Debugf(" load news feed id:%v, news id:%v", feedId, newsId)

	n := news.New(newsId)
	n.FeedId = int64(feedId)
	n.Load()
	log.Info("news:" + n.Title)

	nextNewsId := list.FindNextId(feedId, newsId)
	n.NextId = nextNewsId

	log.Info("show news:", n.Title, ", next:", n.NextId)
	n.MarkRead(0)

	jsonStr, _ := json.Marshal(n)
	w.Write([]byte(jsonStr))
}

type PreviousNewsServer struct {
}

// load previous news
func (server PreviousNewsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	currentNewsId := r.Form.Get("currentId")
	feedId, _ := strconv.Atoi(r.Form.Get("feedId"))
	log.Debugf(" load previous news feed id:%v, news id:%v", feedId, currentNewsId)
	index := list.FindIndexById(feedId, currentNewsId)
	newsIds := list.FindNewsListByRange(list.NewsListKey(feedId), index-1, index-1)
	previousNewsId := newsIds[0]
	n := news.New(previousNewsId)
	n.FeedId = int64(feedId)
	n.Load()
	nextNewsId := list.FindNextId(feedId, previousNewsId)
	n.NextId = nextNewsId

	jsonStr, _ := json.Marshal(n)
	_, _ = w.Write([]byte(jsonStr))
}

type MarkReadServer struct {
}

// 标记整页为已读
func (server MarkReadServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	feedId, _ := strconv.Atoi(r.Form.Get("feedId"))
	log.Debugf("mark this page as read, feed id: %v", feedId)

	readIndex := list.GetLatestReadIndex(userId, feedId)
	// reset read index
	newIndex := readIndex + list.PageSize - 1 //新已读=旧值加每页数量
	count := list.Count(feedId)
	if newIndex > count {
		newIndex = count - 1
	}
	log.Debugf("last read index: %v, new index: %v", readIndex, newIndex)

	list.SetReadIndex(0, feedId, newIndex) //save
	// del read mark set,按feed删除
	news.DelReadMark(0, feedId)

	// load next page
	newsList := loadNewsListByFeed(feedId)
	jsonStr, _ := json.Marshal(newsList)
	_, _ = w.Write([]byte(jsonStr))
}

const port = "3001"

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	log.Info("rssx starting...")

	//同步新闻列表， rss源>redis
	//go rss.Sync()

	//定时清理缓存
	//go rss.Gc()

	dir := config.GetString("ui.path", "")
	log.Info("ui path: ", dir)
	http.Handle("/", http.FileServer(http.Dir(dir)))

	var server HttpServer
	http.Handle("/api/feeds", server)

	var newsListServer NewsListServer
	http.Handle("/api/news-list", newsListServer)

	var newsServer NewsServer
	http.Handle("/api/news", newsServer)

	var previousNewsServer PreviousNewsServer
	http.Handle("/api/news-previous", previousNewsServer)

	var markReadServer MarkReadServer
	http.Handle("/api/mark-read", markReadServer)

	log.Info("rssx server listening:", port)
	err := http.ListenAndServe(":"+port, nil)
	handleErr(err)
}

func handleErr(e error) {
	if e != nil {
		log.Info(e.Error())
	}
}
