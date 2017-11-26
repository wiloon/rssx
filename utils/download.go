package utils

import (
	"net/http"
	"os"
	"io"
)

func Download() {
	url := "http://www.oschina.net/news/rss"
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("osChina.xml")
	if err != nil {
		panic(err)
	}
	io.Copy(f, res.Body)
}
