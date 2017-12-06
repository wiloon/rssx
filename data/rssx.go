package data

import (
	"github.com/wiloon/wiloon-data/mysql"
	"wiloon.com/rssx/feed"
	"github.com/wiloon/wiloon-log/log"
	"wiloon.com/rssx/news"
)

var rssx mysql.Database

func init() {
	config := mysql.Config{DatabaseName: "rssx", Address: "127.0.0.1:3306", Username: "user0", Password: "password0"}
	rssx = mysql.NewDatabase(config)

}

func FindUserFeeds(userId int) []feed.Feed {
	stmt := "select f.id,f.title from user_feed uf join feed f on uf.feed_id=f.id where user_id=?"
	result := rssx.Find(stmt, []interface{}{userId}...)
	var feeds []feed.Feed
	for _, v := range result {
		log.Info(v)

		feeds = append(feeds, feed.Feed{Id: v["id"].(int64), Title: string(v["title"].([]uint8))})

	}
	return feeds
}

func FindNewsByFeed(feedId int) []news.News {
	stmt := "select * from news where feed_id=?"
	result := rssx.Find(stmt, []interface{}{feedId}...)
	var newsList []news.News
	for _, v := range result {
		log.Info(v)

		newsList = append(newsList, news.News{Id: v["news_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url: string(v["url"].([]uint8)),
			Description: string(v["description"].([]uint8)),
		})

	}
	return newsList
}

func SaveNews(newsId, title, url, description string) {
	stmt := "INSERT news SET  feed_id=?,title=?,url=?,description=?"
	rssx.Save(stmt, []interface{}{0, title, url, description}...)
}
