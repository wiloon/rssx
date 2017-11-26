package main

import (
	"net/http"
	"log"
)

type httpServer struct {
}

func (server httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//refresh news

	w.Write([]byte("{\"Status\":\"SUCCESS\"}"))
}

const port = "3000"

func main() {
	log.Println("server starting...")
	http.Handle("/", http.FileServer(http.Dir("/home/wiloon/projects/rssx-client/dist")))
	var server httpServer
	http.Handle("/refresh", server)
	http.ListenAndServe(":"+port, nil)
	log.Println("rssx listening:", port)
}
