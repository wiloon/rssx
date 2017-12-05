package feed

import "wiloon.com/rssx/data"

type Feed struct {
	Id    int
	Title string
	Url   string
}

func (feed *Feed) GetNews() []string {
	return data.Find(0)
}
