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
	"wiloon.com/rssx/rss"
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

	// feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}
	r.ParseForm()
	feedId, _ := strconv.Atoi(r.Form.Get("id"))

	var newsList []news.News
	if feedId == -1 {
		// find all news for all user feeds
		newsList = data.FindAllNewsForUser(userId)
	} else {
		// by feed id
		newsIds := list.FindNewsListByUserFeed(userId, feedId)
		for _, v := range newsIds {
			n := news.New(v)
			n.LoadTitle()
			n.LoadReadFlag(0)

			newsList = append(newsList, *n)

		}
	}

	jsonStr, _ := json.Marshal(newsList)

	w.Write([]byte(jsonStr))
}

type NewsServer struct {
}

var cachedNextNews news.News

// show news
func (server NewsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("load news")
	// feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}
	r.ParseForm()
	newsId := r.Form.Get("id")
	feedId, _ := strconv.Atoi(r.Form.Get("feedId"))
	log.Debugf("feed id:%v, news id:%v", feedId, newsId)

	feed.NewFeed(feedId)

	thisNews := loadNews(feedId, newsId)

	log.Info("show news:", thisNews.Title, ", next:", thisNews.NextId)
	thisNews.MarkRead(0)

	jsonStr, _ := json.Marshal(thisNews)
	w.Write([]byte(jsonStr))

}

func loadNews(feedId int, newsId string) news.News {
	log.Info("find news:" + newsId)
	thisNews := news.New(newsId)
	thisNews.Load()
	log.Info("news:" + thisNews.Title)

	nextNewsId := list.FindNextId(feedId, newsId)
	thisNews.NextId = nextNewsId
	//next := news.News{}
	//if feedId == -1 {
	//	next = data.FindNextNews(userId, newsId)
	//	thisNews.FeedId = -1
	//}
	//thisNews.NextId = next.Id

	return *thisNews
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
