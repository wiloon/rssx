package list

import (
	log "github.com/sirupsen/logrus"
	"rssx/feed"
	"rssx/storage/redisx"
	"strconv"
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

// 新文章 ， 加入 到id集合
// score : 当前时间戳
func (newsList *NewsList) AppendNews(score int64, newsId string) {
	feedNewsKey := FeedNewsKeyPrefix + strconv.Itoa(int(newsList.feed.Id))
	_, _ = redisx.GetConn().Do("ZADD", feedNewsKey, score, newsId)
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

	result, err := redisx.GetConn().Do("ZRANGE", key, start, end)
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
	index := FindIndexById(feedId, newsId)
	nextIndex := index + 1
	foo, _ := redisx.GetConn().Do("ZRANGE", feedNewsKey(feedId), nextIndex, nextIndex)
	if len(foo.([]interface{})) > 0 {
		nextNewsId = string(foo.([]interface{})[0].([]byte))

	} else {
		nextNewsId = ""
	}
	return nextNewsId
}

func feedNewsKey(feedId int) string {
	key := FeedNewsKeyPrefix + strconv.Itoa(feedId)
	log.Debugf("feed news key: %v", key)
	return key
}

// news list read index, value=sorted set range index, not score
const userFeedLatestReadIndex string = "read_index:"

/* todo, 按score取index
redis里保存 score, 取位置时先取score再用score取member,再用member取位置   -_-!!
*/
func GetLatestReadIndex(userId, feedId int) int64 {
	score := 0
	readMarkKey := userFeedLatestReadIndex + strconv.Itoa(userId) + ":" + strconv.Itoa(feedId)
	r, err := redisx.GetConn().Do("GET", readMarkKey)
	if err != nil {
		log.Info(err.Error())
	}
	if r != nil {
		b := r.([]byte)
		i := string(b)
		score, _ = strconv.Atoi(i)
	}
	//r, _ := redisx.Conn.Do("ZRANGEBYSCORE", score, score)
	rank := redisx.GetIndexByScore(readMarkKey, int64(score))
	log.Debugf("latest read mark score, key: %v, score: %v, rank: %v", readMarkKey, score, rank)
	return rank
}

// todo,存score值
func SetReadIndex(userId, feedId int, index int64) {
	// get score by rank
	//ZRANGE
	//ZSCORE

	key := userFeedLatestReadIndex + strconv.Itoa(userId) + ":" + strconv.Itoa(feedId)
	score := redisx.GetScoreByRank(key, index)

	_, _ = redisx.GetConn().Do("SET", key, score)
	log.Debugf("reset read index, index:%v", index)
}

func FindIndexById(feedId int, newsId string) int64 {
	var index int64
	result, err := redisx.GetConn().Do("ZRANK", feedNewsKey(feedId), newsId)
	if err != nil {
		log.Info(err.Error())
	}
	if result == nil {
		index = -1
	} else {
		index = result.(int64)
	}
	log.Debugf("find index by id: %v, index: %v", newsId, index)
	return index
}

func Count(feedId int) int64 {
	var count int64
	result, err := redisx.GetConn().Do("ZCARD", feedNewsKey(feedId))
	if err != nil {
		log.Info(err.Error())
	}
	if result == nil {
		count = 0
	} else {
		count = result.(int64)
	}
	log.Debugf("feed: %v, news count: %v", feedId, count)
	return count
}
