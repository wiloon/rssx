package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
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

	fmt.Println(v)
	fmt.Println("xml name:", v.XMLName)
	fmt.Println("Version:", v.Version)
	fmt.Println("Description:", v.Description)
	fmt.Println("channel:", v.Chan)
	fmt.Println("channel.title:", v.Chan.Title)
	fmt.Println("channel.link:", v.Chan.Link)
	fmt.Println("channel.items:", v.Chan.Items[0].Title)

}
