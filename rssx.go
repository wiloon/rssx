package main

import (
	"encoding/json"
	"github.com/wiloon/app-config"
	"github.com/wiloon/wiloon-log/log"
	"net/http"
	"strconv"
	"wiloon.com/rssx/data"
	"wiloon.com/rssx/feed"
	"wiloon.com/rssx/feed/news/list"
	"wiloon.com/rssx/news"
)

const userId = 0

type HttpServer struct {
}

// user feed list
func (server HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("load user feed list")
	feeds := []feed.Feed{{Id: -1, Title: "All", Url: ""}}

	feeds = append(feeds, data.FindUserFeeds(userId)...)

	jsonStr, _ := json.Marshal(feeds)
	log.Info("api feeds:", jsonStr)
	w.Write([]byte(jsonStr))
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

	feed.NewFeed(feedId)

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
	w.Write([]byte(jsonStr))
}

type MarkReadServer struct {
}

// mark page as read
func (server MarkReadServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	feedId, _ := strconv.Atoi(r.Form.Get("feedId"))
	log.Debugf("mark read, feed id: %v", feedId)

	readIndex := list.GetLatestReadIndex(userId, feedId)
	// reset read index
	newIndex := readIndex + list.PageSize
	list.SetReadIndex(0, feedId, newIndex)
	log.Debugf("set read index:  %v", newIndex)
	// del read mark set
	news.DelReadMark(0, feedId)

	// load next page
	newsList := loadNewsListByFeed(feedId)
	jsonStr, _ := json.Marshal(newsList)
	w.Write([]byte(jsonStr))
}

const port = "3000"

func main() {
	log.Info("rssx starting...")

	//start rss sync
	//go rss.Sync()

	dir := config.GetString("client.dir", "")
	log.Info("client dir:", dir)
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
