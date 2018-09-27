package redisx

import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
	"wiloon.com/rssx/news"
	"wiloon.com/rssx/utils"
	"wiloon.com/wiloon-log/log"
)

var conn redis.Conn

const feedNewsKeyPrefix string = "feed_news:"
const newsKeyPrefix string = "news:"

func init() {
	var err error
	conn, err = redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Info("failed to connect to redis:" + err.Error())
	}

}
func SaveNews(feedId int64, title, url, description string, pubDate time.Time, guid string) {
	//news := news.News{Guid: guid, Title: title, Url: url, Description: description, PubDate: pubDate.Format("2006-01-02 15:04:05")}
	//newsJson, _ := json.Marshal(news)

	feedNewsKey := feedNewsKeyPrefix + strconv.Itoa(int(feedId))
	score := pubDate.UnixNano() / int64(time.Millisecond)

	newsId := utils.Md5(guid)
	conn.Do("ZADD", feedNewsKey, score, newsId)
	conn.Do("HMSET", "news:"+newsId,
		news.FeedId, feedId,
		news.Title, title,
		news.Url, url,
		news.Description, description,
		news.PubDate, pubDate,
		news.Guid, guid)
	log.Info("save news:" + title)
}

func FindNewsListByUserFeed(userId, feedId int) []news.News {
	var newsList []news.News
	result, err := conn.Do("ZRANGE", feedNewsKeyPrefix+strconv.Itoa(feedId), 0, 9)
	if err != nil {
		log.Info("failed to get news")
	}
	for _, v := range result.([]interface{}) {
		b := v.([]byte)
		newsId := string(b)
		log.Info("news id: " + newsId)
		//n, _ := redis.Values(conn.Do("HGETALL", newsKeyPrefix+newsId))
		n, _ := redis.Values(conn.Do("HMGET", newsKeyPrefix+newsId, news.Title))
		title := string(n[0].([]byte))

		//var foo news.News
		//for _, v := range n {
		//	fmt.Printf("%s ", v.([]byte))
		//}
		//fmt.Printf("\n")
		newsList = append(newsList, news.News{Id: newsId, Title: title})
		log.Info("news list item:" + title)
	}
	log.Info("find news list by feed,news size:" + string(len(newsList)))
	return newsList
}

func FindNews(newsId string) news.News {
	result, err := redis.Values(conn.Do("HMGET", newsKeyPrefix+newsId, news.Title, news.Url, news.Description, news.NextNewsId))
	if err != nil {
		log.Info(err.Error())
	}
	return news.News{Id: newsId,
		Title:       string(result[0].([]byte)),
		Url:         string(result[1].([]byte)),
		Description: string(result[2].([]byte)),

	}
}
