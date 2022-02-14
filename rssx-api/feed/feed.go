package feed

type Feed struct {
	Id    int64
	Title string
	Url   string
}

func NewFeed(feedId int) *Feed {
	var result = new(Feed)
	result.Id = int64(feedId)
	return result
}
