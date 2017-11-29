package main

import (
	"net/http"
	"log"
	"wiloon.com/rssx/feed"
	"encoding/json"
)

type httpServer struct {
}

func (server httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//refresh news
	feed := feed.Feed{}
	list := feed.GetNews()
	jsonstr, _ := json.Marshal(list)

	w.Write([]byte(jsonstr))
}

const port = "3000"

func main() {
	log.Println("server starting...")
	http.Handle("/", http.FileServer(http.Dir("/home/roy/my-projects/rssx-client/dist")))
	var server httpServer
	http.Handle("/refresh", server)
	http.ListenAndServe(":"+port, nil)
	log.Println("rssx listening:", port)
}
