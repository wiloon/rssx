package rss

import "encoding/xml"

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
	Guid        string   `xml:"guid"`
	PubDate     string   `xml:"pubDate"`
}
