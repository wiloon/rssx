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

	feeds := []feed.Feed{feed.Feed{Id: 0, Title: "t0", Url: "u0"}, feed.Feed{Id: 1, Title: "t1", Url: "u1"}}

	jsonStr, _ := json.Marshal(feeds)

	w.Write([]byte(jsonStr))
}

const port = "3000"

func main() {
	log.Println("server starting...")
	http.Handle("/", http.FileServer(http.Dir("/home/wiloon/projects/rssx-client/dist")))

	var server httpServer
	http.Handle("/api/feeds", server)
	log.Println("rssx listening:", port)

	http.ListenAndServe(":"+port, nil)

}
