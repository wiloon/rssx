package list

import (
	"github.com/wiloon/wiloon-log/log"
	"strconv"
	"wiloon.com/rssx/feed"
	"wiloon.com/rssx/storage/redisx"
)

const FeedNewsKeyPrefix string = "feed_news:"
const PageSize int64 = 10

type NewsList struct {
	userId int
	feed   feed.Feed
}

func NewList(userId int, feed feed.Feed) *NewsList {
	var result = new(NewsList)
	result.userId = userId
	result.feed = feed
	return result
}
func (newsList *NewsList) AppendNews(score int64, newsId string) {
	feedNewsKey := FeedNewsKeyPrefix + strconv.Itoa(int(newsList.feed.Id))
	redisx.Conn.Do("ZADD", feedNewsKey, score, newsId)
}

func FindNewsListByUserFeed(userId, feedId int) []string {
	var newsList []string

	latestReadIndex := GetLatestReadIndex(userId, feedId)
	key := NewsListKey(feedId)

	newsList = FindNewsListByRange(key, latestReadIndex, latestReadIndex+PageSize-1)
	log.Infof("find news list by feed,read index: %v, news size: %v", latestReadIndex, len(newsList))
	return newsList
}

func NewsListKey(feedId int) string {
	return FeedNewsKeyPrefix + strconv.Itoa(feedId)
}

func FindNewsListByRange(key string, start, end int64) []string {
	var newsidList []string

	result, err := redisx.Conn.Do("ZRANGE", key, start, end)
	if err != nil {
		log.Info("failed to get news")
	}
	for _, v := range result.([]interface{}) {
		b := v.([]byte)
		newsId := string(b)
		log.Info("news id: " + newsId)
		newsidList = append(newsidList, newsId)
	}
	return newsidList
}

func FindNextId(feedId int, newsId string) string {
	var nextNewsId string
	result, err := redisx.Conn.Do("ZRANK", FeedNewsKeyPrefix+strconv.Itoa(feedId), newsId)
	if err != nil {
		log.Info(err.Error())
	}

	nextIndex := result.(int64) + 1
	foo, _ := redisx.Conn.Do("ZRANGE", FeedNewsKeyPrefix+strconv.Itoa(feedId), nextIndex, nextIndex)
	nextNewsId = string(foo.([]interface{})[0].([]byte))

	return nextNewsId
}

// news list read index, value=sorted set range index, not score
const userFeedLatestReadIndex string = "read_index:"

func GetLatestReadIndex(userId, feedId int) int64 {
	result := 0
	readIndexKey := userFeedLatestReadIndex + strconv.Itoa(userId) + ":" + strconv.Itoa(feedId)
	r, err := redisx.Conn.Do("GET", readIndexKey)
	if err != nil {
		log.Info(err.Error())
	}
	if r != nil {
		b := r.([]byte)
		i := string(b)
		result, _ = strconv.Atoi(i)
	}
	log.Debugf("latest read index: %v", result)
	return int64(result)
}

func SetReadIndex(userId, feedId int, score int64) {

	redisx.Conn.Do("SET", userFeedLatestReadIndex+strconv.Itoa(userId)+":"+strconv.Itoa(feedId), score)
	log.Debugf("reset read index, index:%v", score)
}

func FindIndexById(feedId int, newsId string) int64 {
	var index int64
	result, err := redisx.Conn.Do("ZRANK", FeedNewsKeyPrefix+strconv.Itoa(feedId), newsId)
	if err != nil {
		log.Info(err.Error())
	}

	index = result.(int64)

	return index
}
