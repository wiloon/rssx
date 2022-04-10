package news

import (
	"github.com/garyburd/redigo/redis"
	"rssx/storage/redisx"
	"rssx/user"
	log "rssx/utils/logger"
	"strconv"
)

func init() {

}

const (
	FeedId      = "FeedId"
	Title       = "Title"
	Url         = "Url"
	Description = "Description"
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
	_, _ = redisx.Exec("HMSET", "news:"+n.Id,
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

func (n *News) IsRead(userId string) bool {
	read := false
	readMarkKey := newsReadMark + userId + ":" + strconv.Itoa(int(n.FeedId))
	log.Debugf("check news is read, read flag key: %v, news id: %v", readMarkKey, n.Id)
	r, _ := redisx.Exec("SISMEMBER", readMarkKey, n.Id)
	if r != nil && r.(int64) == 1 {
		read = true
	}
	log.Debugf("check news is read, read flag key: %v, news id: %v, read flag :%v", readMarkKey, n.Id, read)
	return read
}

func (n *News) MarkRead(userId int) {
	_, _ = redisx.Exec("SADD", newsReadMark+strconv.Itoa(userId)+":"+strconv.Itoa(int(n.FeedId)), n.Id)
	log.Debugf("mark news as read, news id: %v", n.Id)
}

const newsKeyPrefix string = "news:"

func (n *News) LoadTitle() {
	result, _ := redis.Values(redisx.Exec("HMGET", newsKeyPrefix+n.Id, Title))
	if result != nil && len(result) > 0 {
		n.Title = string(result[0].([]byte))
	}
}
func (n *News) LoadReadFlag() {
	n.ReadFlag = n.IsRead(user.DefaultId)
	log.Debugf("read mark, news id: %v, title: %v", n.Id, n.Title)
}
func (n *News) Load() {
	result, err := redis.Values(redisx.Exec("HMGET", newsKeyPrefix+n.Id, Title, Url, Description, Score, PubDate))
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
	_, _ = redisx.Exec("DEL", newsReadMark+strconv.Itoa(userId)+":"+strconv.Itoa(int(feedId)))
}
