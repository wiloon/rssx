package list

import (
	"strconv"
	"wiloon.com/rssx/feed"
	"wiloon.com/rssx/storage/redisx"
	"wiloon.com/wiloon-log/log"
)

const FeedNewsKeyPrefix string = "feed_news:"

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
	result, err := redisx.Conn.Do("ZRANGE", FeedNewsKeyPrefix+strconv.Itoa(feedId), 0, 9)
	if err != nil {
		log.Info("failed to get news")
	}
	for _, v := range result.([]interface{}) {
		b := v.([]byte)
		newsId := string(b)
		log.Info("news id: " + newsId)
		//n, _ := redis.Values(Conn.Do("HGETALL", newsKeyPrefix+newsId))
		//n, _ := redis.Values(redisx.Conn.Do("HMGET", newsKeyPrefix+newsId, news.Title))
		//title := string(n[0].([]byte))

		//var foo news.News
		//for _, v := range n {
		//	fmt.Printf("%s ", v.([]byte))
		//}
		//fmt.Printf("\n")
		newsList = append(newsList, newsId)
		//log.Info("news list item:" + title)
	}
	log.Info("find news list by feed,news size:" + string(len(newsList)))
	return newsList
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
