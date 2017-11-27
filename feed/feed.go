package feed

import "wiloon.com/rssx/data"

type Feed struct {
	feedId int
	title  string
	url    string
}

func (feed *Feed) GetNews() {
	data.Find(0)
}
