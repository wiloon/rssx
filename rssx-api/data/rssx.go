package data

import (
	"rssx/feed"
	"rssx/utils/config"
	log "rssx/utils/logger"
	"rssx/utils/mysql"
)

var rssx *mysql.Database

func init() {

}
func Rssx() *mysql.Database {
	if rssx == nil {
		address := config.GetString("mysql.address", "127.0.0.1:3306")
		user := config.GetString("mysql.user", "user0")
		password := config.GetString("mysql.password", "password0")
		mysqlConfig := mysql.Config{DatabaseName: "rssx", Address: address, Username: user, Password: password}
		rssx = mysql.NewDatabase(mysqlConfig)
	}
	return rssx
}
func FindUserFeeds(userId int) []feed.Feed {
	stmt := "select f.feed_id,f.title from user_feed uf join feed f on uf.feed_id=f.feed_id where user_id=?"

	result := Rssx().Find(stmt, []interface{}{userId}...)
	var feeds []feed.Feed
	for _, v := range result {
		feeds = append(feeds, feed.Feed{Id: v["feed_id"].(int64), Title: string(v["title"].([]uint8))})
	}
	log.Infof("user feeds size: %v", len(feeds))
	return feeds
}

func FindFeeds() []feed.Feed {
	stmt := "select feed_id,title,url from feed where deleted=?"
	result := Rssx().Find(stmt, []interface{}{0}...)
	var feeds []feed.Feed
	for _, v := range result {
		feeds = append(feeds, feed.Feed{
			Id:    v["feed_id"].(int64),
			Title: string(v["title"].([]uint8)),
			Url:   string(v["url"].([]uint8)),
		})
	}
	return feeds
}
