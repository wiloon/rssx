package main

import (
	"encoding/json"
	"github.com/wiloon/wiloon-log/log"
	"net/http"
	"strconv"
	"wiloon.com/rssx/data"
	"wiloon.com/rssx/feed"
	"wiloon.com/rssx/news"

	"github.com/wiloon/app-config"
	"wiloon.com/rssx/rss"
)

const userId = 0

type HttpServer struct {
}

// user feeds
func (server HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	feeds := []feed.Feed{{Id: -1, Title: "All", Url: ""}}

	feeds = append(feeds, data.FindUserFeeds(userId)...)

	jsonStr, _ := json.Marshal(feeds)
	log.Info("api feeds:", jsonStr)
	w.Write([]byte(jsonStr))
}

type NewsListServer struct {
}

// news list by feed
func (server NewsListServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}
	r.ParseForm();
	feedId, _ := strconv.Atoi(r.Form.Get("id"))

	var newsList []news.News
	if feedId == -1 {
		// find all news for all user feeds
		newsList = data.FindAllNewsForUser(userId)
	} else {
		// by feed id
		newsList = data.FindNewsListByUserFeed(userId, feedId)
	}

	jsonStr, _ := json.Marshal(newsList)

	w.Write([]byte(jsonStr))
}

type NewsServer struct {
}

var cachedNextNews news.News

// show news
func (server NewsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}
	r.ParseForm();
	newsId, _ := strconv.Atoi(r.Form.Get("id"))
	feedId, _ := strconv.Atoi(r.Form.Get("feedId"))

	newsId64 := int64(newsId)

	thisNews := news.News{}
	if cachedNextNews.Id == newsId64 {
		thisNews = cachedNextNews
	} else {
		thisNews = loadNews(feedId, newsId64)
	}

	log.Info("show news:", thisNews.Title, ", next:", thisNews.NextId)

	//mark  as read
	data.MarkNewsRead(userId, thisNews.Id)
	jsonStr, _ := json.Marshal(thisNews)
	w.Write([]byte(jsonStr))

	go loadNextNews(feedId, thisNews.NextId)

}

func loadNextNews(feedId int, nextNewsId int64) {
	cachedNextNews = loadNews(feedId, nextNewsId)
}

func loadNews(feedId int, newsId int64) news.News {
	thisNews := data.FindNews(newsId)
	next := news.News{}
	if feedId == -1 {
		next = data.FindNextNews(userId, newsId)
		thisNews.FeedId = -1
	} else {
		next = data.FindNextNewsByFeed(userId, feedId, newsId)
	}
	thisNews.NextId = next.Id

	return thisNews
}

const port = "3000"

func main() {
	log.Info("rssx starting...")

	//start rss sync
	go rss.Sync()

	dir := config.GetString("client.dir", "")
	log.Info("client dir:", dir)
	http.Handle("/", http.FileServer(http.Dir(dir)))

	var server HttpServer
	http.Handle("/api/feeds", server)

	var newsListServer NewsListServer
	http.Handle("/api/news-list", newsListServer)

	var newsServer NewsServer
	http.Handle("/api/news", newsServer)
	log.Info("rssx listening:", port)

	http.ListenAndServe(":"+port, nil)
}
