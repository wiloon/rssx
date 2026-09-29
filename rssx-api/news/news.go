package news

import (
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"rssx/storage/redisx"
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

// ReadMarkKey is the Redis set of articles one user has read out of order in
// one feed.
func ReadMarkKey(userId string, feedId int64) string {
	return newsReadMark + userId + ":" + strconv.FormatInt(feedId, 10)
}

func (n *News) IsRead(userId string) bool {
	read := false
	readMarkKey := ReadMarkKey(userId, n.FeedId)
	log.Debugf("check news is read, read flag key: %v, news id: %v", readMarkKey, n.Id)
	r, _ := redisx.Exec("SISMEMBER", readMarkKey, n.Id)
	if r != nil && r.(int64) == 1 {
		read = true
	}
	log.Debugf("check news is read, read flag key: %v, news id: %v, read flag :%v", readMarkKey, n.Id, read)
	return read
}

func (n *News) MarkRead(userId string) {
	_, _ = redisx.Exec("SADD", ReadMarkKey(userId, n.FeedId), n.Id)
	log.Debugf("mark news as read, user id: %v, news id: %v", userId, n.Id)
}

const newsKeyPrefix string = "news:"

func (n *News) LoadTitle() {
	result, _ := redis.Values(redisx.Exec("HMGET", newsKeyPrefix+n.Id, Title))
	if result != nil && len(result) > 0 {
		n.Title = string(result[0].([]byte))
	}
}

// ErrNewsNotFound is returned by Load when the article hash does not exist,
// e.g. it was removed by GC or the id is unknown.
var ErrNewsNotFound = errors.New("news not found")

// Load fills the article's fields from its Redis hash.
func (n *News) Load() error {
	fields, err := redis.Values(redisx.Exec("HMGET", newsKeyPrefix+n.Id, Title, Url, Description, Score, PubDate))
	if err != nil {
		return fmt.Errorf("load news %s: %w", n.Id, err)
	}
	if len(fields) != 5 || fields[0] == nil {
		return ErrNewsNotFound
	}
	values, err := redis.Strings(fields, nil)
	if err != nil {
		return fmt.Errorf("load news %s: %w", n.Id, err)
	}

	n.Title = values[0]
	n.Url = values[1]
	n.Description = values[2]
	n.Score, _ = strconv.ParseInt(values[3], 10, 64)
	n.PubDate = values[4]
	return nil
}

func DelReadMark(userId string, feedId int) {
	_, _ = redisx.Exec("DEL", ReadMarkKey(userId, int64(feedId)))
}

// LoadListForFeed loads titles and read flags for a page of news in one feed
// using two pipelined Redis round trips total, instead of two per item.
func LoadListForFeed(feedId int64, userId string, ids []string) []News {
	result := make([]News, 0, len(ids))
	if len(ids) == 0 {
		return result
	}

	titles := make([]string, len(ids))
	readFlags := make([]bool, len(ids))
	readMarkKey := ReadMarkKey(userId, feedId)

	err := redisx.WithConn(func(conn redis.Conn) error {
		for _, id := range ids {
			_ = conn.Send("HGET", newsKeyPrefix+id, Title)
			_ = conn.Send("SISMEMBER", readMarkKey, id)
		}
		if err := conn.Flush(); err != nil {
			return err
		}
		for i := range ids {
			title, _ := redis.String(conn.Receive())
			titles[i] = title
			isRead, _ := redis.Bool(conn.Receive())
			readFlags[i] = isRead
		}
		return nil
	})
	if err != nil {
		log.Errorf("load news list for feed failed, feed id: %v, err: %v", feedId, err)
		return result
	}

	for i, id := range ids {
		result = append(result, News{
			Id:       id,
			FeedId:   feedId,
			Title:    titles[i],
			ReadFlag: readFlags[i],
		})
	}
	return result
}
