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

func FindAllNewsForUser(userId int) []news.News {
	stmt := `
SELECT n.news_id,n.title,n.url,n.description,n.feed_id
FROM news n
JOIN user_feed uf ON n.feed_id=uf.feed_id
LEFT JOIN news_read_mark nrm ON n.news_id = nrm.news_id
WHERE uf.user_id=? AND nrm.news_id IS NULL
ORDER BY n.news_id
`

	result := rssx.Find(stmt, []interface{}{userId}...)
	var newsList []news.News
	for _, v := range result {
		newsList = append(newsList, news.News{Id: v["news_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url: string(v["url"].([]uint8)),
			Description: string(v["description"].([]uint8)),
			FeedId: v["feed_id"].(int64),
		})
	}
	return newsList
}

func FindNewsListByUserFeed(userId, feedId int) []news.News {
	stmt := `
SELECT nrm.news_id AS unread_news_id,n.news_id,n.title,n.url,n.description,n.feed_id from news n
JOIN user_feed uf ON uf.user_id=? and uf.feed_id=? AND n.feed_id=uf.feed_id
LEFT JOIN news_read_mark nrm ON nrm.user_id=? and n.news_id=nrm.news_id
where nrm.news_id IS NULL
ORDER BY n.news_id
`

	result := rssx.Find(stmt, []interface{}{userId, feedId, userId}...)
	var newsList []news.News
	for _, v := range result {
		newsList = append(newsList, news.News{Id: v["news_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url: string(v["url"].([]uint8)),
			Description: string(v["description"].([]uint8)),
			FeedId: v["feed_id"].(int64),
		})
	}
	return newsList
}

func FindNextNewsByFeed(userId, feedId int, newsId int) news.News {
	var newsRtn news.News

	stmt := `
SELECT n.news_id
FROM news n
JOIN user_feed uf ON n.feed_id=uf.feed_id
LEFT JOIN news_read_mark nrm ON n.news_id = nrm.news_id
WHERE uf.user_id=? AND nrm.news_id IS NULL AND n.news_id!=? and uf.feed_id=?
ORDER BY n.news_id  limit 1

`
	result := rssx.Find(stmt, []interface{}{userId, newsId, feedId}...)
	var newsList []news.News
	for _, v := range result {
		newsList = append(newsList, news.News{
			Id: v["news_id"].(int64),
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

func FindNews(newsId int) news.News {
	var newsRtn news.News

	stmt := `
select * from news where news_id=?
`
	result := rssx.Find(stmt, []interface{}{newsId}...)
	var newsList []news.News
	for _, v := range result {
		newsList = append(newsList, news.News{Id: v["news_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url: string(v["url"].([]uint8)),
			Description: string(v["description"].([]uint8)),
			FeedId: v["feed_id"].(int64),
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

func FindNextNews(userId, newsId int) news.News {
	var newsRtn news.News

	stmt := `
SELECT n.news_id
FROM news n
JOIN user_feed uf ON n.feed_id=uf.feed_id
LEFT JOIN news_read_mark nrm ON n.news_id = nrm.news_id
WHERE uf.user_id=? AND nrm.news_id IS NULL AND n.news_id!=?
ORDER BY n.news_id limit 1

`
	result := rssx.Find(stmt, []interface{}{userId, newsId}...)
	var newsList []news.News
	for _, v := range result {
		newsList = append(newsList, news.News{
			Id: v["news_id"].(int64),
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
	stmt := "INSERT news_read_mark SET  user_id=?,news_id=?"
	rssx.Save(stmt, []interface{}{userId, newsId}...)
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
