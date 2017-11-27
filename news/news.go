package news

import (
	"wiloon.com/rssx/data"
)

func init() {

}

var bucket = "NewsBucket"

type Site struct {
	Title    string
	NewsList []News
}
type News struct {
	Title       string
	Url         string
	Description string
}

func (site *Site) Append(title, url, description string) {
	site.NewsList = append(site.NewsList, News{Title: title, Url: url, Description: description})
}
func (site *Site) Save() {

	for _, v := range site.NewsList {
		data.Save(v.Title, v.Url, v.Description)
	}

}
