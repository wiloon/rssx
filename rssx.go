package main

import (
	"net/http"
	"log"
	"encoding/json"

	"wiloon.com/rssx/data"
	"strconv"
)

type httpServer struct {
}

func (server httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}
	feeds := data.FindUserFeeds(0)

	jsonStr, _ := json.Marshal(feeds)

	w.Write([]byte(jsonStr))
}

type NewsListServer struct {
}

func (server NewsListServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}
	r.ParseForm();
	id, _ := strconv.Atoi(r.Form.Get("id"))
	newsList := data.FindNewsListByFeed(id)

	jsonStr, _ := json.Marshal(newsList)

	w.Write([]byte(jsonStr))
}

type NewsServer struct {
}

func (server NewsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}
	r.ParseForm();
	id, _ := strconv.Atoi(r.Form.Get("id"))
	news := data.FindNews(id)

	jsonStr, _ := json.Marshal(news)
	w.Write([]byte(jsonStr))
}

const port = "3000"

func main() {
	log.Println("server starting...")
	http.Handle("/", http.FileServer(http.Dir("/home/roy/my-projects/rssx-client/dist")))

	var server httpServer
	http.Handle("/api/feeds", server)
	var newsListServer NewsListServer
	http.Handle("/api/news-list", newsListServer)

	var newsServer NewsServer
	http.Handle("/api/news", newsServer)
	log.Println("rssx listening:", port)

	http.ListenAndServe(":"+port, nil)
}
