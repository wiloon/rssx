package news

import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"wiloon.com/rssx/storage/redisx"
	"wiloon.com/wiloon-log/log"
)

func init() {

}

var bucket = "NewsBucket"

const (
	NewsId      = "Id"
	FeedId      = "FeedId"
	Title       = "Title"
	Url         = "Url"
	Description = "Description"
	NextNewsId  = "NextId"
	PubDate     = "PubDate"
	Guid        = "Guid"
	Score       = "Score"
)

type Site struct {
	Title    string
	NewsList []News
}
type News struct {
	Id          string
	FeedId      int64
	Title       string
	Url         string
	Description string
	NextId      string
	PubDate     string
	Guid        string
	Score       int64
	UnRead      bool
}

func New(newsId string) *News {
	var result = new(News)
	result.Id = newsId
	return result

}
func (site *Site) Append(title, url, description string) {
	site.NewsList = append(site.NewsList, News{Title: title, Url: url, Description: description})
}

func (n *News) Save() {

	redisx.Conn.Do("HMSET", "news:"+n.Id,
		FeedId, n.FeedId,
		Title, n.Title,
		Url, n.Url,
		Description, n.Description,
		PubDate, n.PubDate,
		Guid, n.Guid,
		Score, n.Score,
	)

	log.Info("save news:" + n.Title)
}

const newsReadMark string = "feed_read:"

func (n *News) IsRead(userId int) bool {
	r, _ := redisx.Conn.Do("SISMEMBER", newsReadMark+strconv.Itoa(userId)+":"+strconv.Itoa(int(n.FeedId)), n.Id)
	return r == 1
}

func (n *News) MarkRead(userId int) {
	redisx.Conn.Do("SADD", newsReadMark+strconv.Itoa(userId)+":"+strconv.Itoa(int(n.FeedId)), n.Id)
}

const newsKeyPrefix string = "news:"

func (n *News) LoadTitle() {
	result, _ := redis.Values(redisx.Conn.Do("HMGET", newsKeyPrefix+n.Id, Title))
	n.Title = string(result[0].([]byte))
}
func (n *News) LoadReadFlag(userId int) {
	n.UnRead = !n.IsRead(userId)

}
func (n *News) Load() {
	result, err := redis.Values(redisx.Conn.Do("HMGET", newsKeyPrefix+n.Id, Title, Url, Description, Score))
	if err != nil {
		log.Info(err.Error())
	}

	n.Title = string(result[0].([]byte))
	n.Url = string(result[1].([]byte))
	n.Description = string(result[2].([]byte))
	score, _ := strconv.Atoi(string(result[3].([]byte)))
	n.Score = int64(score)

}
