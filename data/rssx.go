package data

import (
	"github.com/wiloon/wiloon-data/mysql"
	"wiloon.com/rssx/feed"
	"github.com/wiloon/wiloon-log/log"
	"wiloon.com/rssx/news"
	"time"
)

var rssx mysql.Database

func init() {
	config := mysql.Config{DatabaseName: "rssx", Address: "127.0.0.1:3306", Username: "user0", Password: "password0"}
	rssx = mysql.NewDatabase(config)

}

func FindUserFeeds(userId int) []feed.Feed {
	stmt := "select f.feed_id,f.title from user_feed uf join feed f on uf.feed_id=f.feed_id where user_id=?"

	result := rssx.Find(stmt, []interface{}{userId}...)
	var feeds []feed.Feed
	for _, v := range result {
		log.Info(v)

		feeds = append(feeds, feed.Feed{Id: v["feed_id"].(int64), Title: string(v["title"].([]uint8))})

	}
	return feeds
}

func FindNewsListByFeed(feedId int) []news.News {
	stmt := `select un.news_id as unread_news_id,n.news_id,n.title,n.url,n.description
from news n left join user_news un on n.news_id=un.news_id
where un.news_id is null and n.feed_id=? order by news_id`

	result := rssx.Find(stmt, []interface{}{feedId}...)
	var newsList []news.News
	for _, v := range result {
		newsList = append(newsList, news.News{Id: v["news_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url: string(v["url"].([]uint8)),
			Description: string(v["description"].([]uint8)),
		})
	}
	return newsList
}

func FindNews(newsId int) news.News {
	var newsRtn news.News

	stmt := "select * from news where news_id>=? order by news_id limit 2"
	result := rssx.Find(stmt, []interface{}{newsId}...)
	var newsList []news.News
	for _, v := range result {
		newsList = append(newsList, news.News{Id: v["news_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url: string(v["url"].([]uint8)),
			Description: string(v["description"].([]uint8)),
		})
	}
	if len(newsList) > 1 {
		newsList[0].NextId = newsList[1].Id
	}

	if len(newsList) > 0 {
		newsRtn = newsList[0]
	}
	return newsRtn
}

func SaveNews(feedId int64, title, url, description string, pubDate time.Time, guid string) {
	stmt := "INSERT news SET  feed_id=?,title=?,url=?,description=?,pub_date=?,guid=?"
	rssx.Save(stmt, []interface{}{feedId, title, url, description, pubDate, guid}...)
}

func FindFeeds() []feed.Feed {
	stmt := "select feed_id,title,url from feed where deleted=?"
	result := rssx.Find(stmt, []interface{}{0}...)
	var feeds []feed.Feed
	for _, v := range result {
		log.Info(v)

		feeds = append(feeds, feed.Feed{
			Id:    v["feed_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url:   string(v["url"].([]uint8)),
		})

	}
	return feeds
}

func MarkNewsRead(userId, newsId int) {
	stmt := "INSERT user_news SET  user_id=?,news_id=?,read_mark=?"
	rssx.Save(stmt, []interface{}{userId, newsId, 1}...)
}

func FindLatestNewsByFeed(feedId int64) news.News {
	stmt := "select * from news where feed_id=? order by pub_date desc limit 1"
	result := rssx.Find(stmt, []interface{}{feedId}...)
	var newsList []news.News
	for _, v := range result {

		log.Info(v)
		newsList = append(newsList, news.News{Id: v["news_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url: string(v["url"].([]uint8)),
			Description: string(v["description"].([]uint8)),
			PubDate: string(v["pub_date"].([]uint8)),
		})

	}

	return newsList[0]
}

func FindNewsByGuid(guid string) []news.News {
	stmt := "select news_id from news where guid=?"
	result := rssx.Find(stmt, []interface{}{guid}...)
	var newsList []news.News
	for _, v := range result {

		log.Info(v)
		newsList = append(newsList, news.News{Id: v["news_id"].(int64)})
	}

	return newsList
}
