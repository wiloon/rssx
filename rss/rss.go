package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"log"

	"wiloon.com/rssx/news"
)

type Rss struct {
	XMLName     xml.Name `xml:"rss"`
	Version     string   `xml:"version,attr"`
	Chan        channel  `xml:"channel"`
	Description string   `xml:",innerxml"`
}

type channel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Language    string   `xml:"language"`
	PubDate     string   `xml:"pubDate"`
	Items       []item   `xml:"item"`
}

type item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Category    string   `xml:"category"`
	Description string   `xml:"description"`
}

func main() {
	file, err := os.Open("/home/roy/gopath/src/wiloon.com/rssx/osChina.xml") // For read access.
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	v := Rss{}
	err = xml.Unmarshal(data, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	log.Println(v)
	log.Println("xml name:", v.XMLName)
	log.Println("Version:", v.Version)
	log.Println("Description:", v.Description)
	log.Println("channel:", v.Chan)
	log.Println("channel.title:", v.Chan.Title)
	log.Println("channel.link:", v.Chan.Link)
	log.Println("channel.items:", v.Chan.Items[0].Title)
	log.Println("channel.items:", v.Chan.Items[1].Title)
	osChina := news.Site{Title: "osChina"}
	for i, v := range v.Chan.Items {
		log.Printf("index:%v, title:%v", i, v.Title)
		osChina.Append(v.Title, v.Link, v.Description)
	}
	osChina.Save()
}
