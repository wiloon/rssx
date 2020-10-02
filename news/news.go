package news

import (
	"github.com/garyburd/redigo/redis"
	log "github.com/wiloon/pingd-log/logconfig/zaplog"
	"rssx/storage/redisx"
	"strconv"
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
	ReadFlag    bool
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
	_, _ = redisx.GetConn().Do("HMSET", "news:"+n.Id,
		FeedId, n.FeedId,
		Title, n.Title,
		Url, n.Url,
		Description, n.Description,
		PubDate, n.PubDate,
		Guid, n.Guid,
		Score, n.Score,
	)
	log.Debug("save news:" + n.Title)
}

// read mark, redis set, value=news id
const newsReadMark string = "read_mark:"

func (n *News) IsRead(userId int) bool {
	read := false
	readMarkKey := newsReadMark + strconv.Itoa(userId) + ":" + strconv.Itoa(int(n.FeedId))
	r, _ := redisx.GetConn().Do("SISMEMBER", readMarkKey, n.Id)
	if r.(int64) == 1 {
		read = true
	}
	log.Debugf("check news is read, read flag key: %v, news id: %v, read flag :%v", readMarkKey, n.Id, read)
	return read
}

func (n *News) MarkRead(userId int) {
	_, _ = redisx.GetConn().Do("SADD", newsReadMark+strconv.Itoa(userId)+":"+strconv.Itoa(int(n.FeedId)), n.Id)
	log.Debugf("mark news as read, news id: %v", n.Id)
}

const newsKeyPrefix string = "news:"

func (n *News) LoadTitle() {
	result, _ := redis.Values(redisx.GetConn().Do("HMGET", newsKeyPrefix+n.Id, Title))
	n.Title = string(result[0].([]byte))
}
func (n *News) LoadReadFlag(userId int) {

	n.ReadFlag = n.IsRead(0)

	log.Debugf("read mark, news id: %v, title: %v", n.Id, n.Title)
}
func (n *News) Load() {
	result, err := redis.Values(redisx.GetConn().Do("HMGET", newsKeyPrefix+n.Id, Title, Url, Description, Score, PubDate))
	if err != nil {
		log.Info(err.Error())
	}

	n.Title = string(result[0].([]byte))
	n.Url = string(result[1].([]byte))
	n.Description = string(result[2].([]byte))
	score, _ := strconv.Atoi(string(result[3].([]byte)))
	n.Score = int64(score)
	n.PubDate = string(result[4].([]byte))

}

func DelReadMark(userId, feedId int) {
	_, _ = redisx.GetConn().Do("DEL", newsReadMark+strconv.Itoa(userId)+":"+strconv.Itoa(int(feedId)))
}
