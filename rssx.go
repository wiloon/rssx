package main

import (
	"net/http"
	"encoding/json"
	"wiloon.com/rssx/data"
	"strconv"
	"github.com/wiloon/wiloon-log/log"
	//"wiloon.com/rssx/rss"
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
	id, _ := strconv.Atoi(r.Form.Get("id"))

	var newsList []news.News
	if id == -1 {
		// find all news for all user feeds
		newsList = data.FindAllNewsForUser(userId)
	} else {
		// by feed id
		newsList = data.FindNewsListByUserFeed(userId, id)
	}

	jsonStr, _ := json.Marshal(newsList)

	w.Write([]byte(jsonStr))
}

type NewsServer struct {
}

// show news
func (server NewsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}
	r.ParseForm();
	newsId, _ := strconv.Atoi(r.Form.Get("id"))
	feedId, _ := strconv.Atoi(r.Form.Get("feedId"))

	thisNews := data.FindNews(newsId)

	if thisNews.Id != 0 {
		next := news.News{}
		if feedId == -1 {
			next = data.FindNextNews(userId, newsId)
		} else {
			next = data.FindNextNewsByFeed(userId, feedId, newsId)
		}
		thisNews.NextId = next.Id
		log.Info("show news:", thisNews.Title, ", next:", thisNews.NextId)

		//mark  as read
		data.MarkNewsRead(userId, newsId)
		jsonStr, _ := json.Marshal(thisNews)
		w.Write([]byte(jsonStr))
	}

}

const port = "3000"

func main() {
	log.Info("server starting...")

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
