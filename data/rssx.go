package data

import (
	"wiloon.com/wiloon-data/mysql"
	"wiloon.com/rssx/feed"
	"github.com/wiloon/wiloon-log/log"
)

var rssx mysql.Database

func init() {
	config := mysql.Config{DatabaseName: "rssx", Address: "127.0.0.1:3306", Username: "user0", Password: "password0"}
	rssx = mysql.NewDatabase(config)

}

func FindUserFeeds(userId int) []feed.Feed {
	stmt := "select f.id,f.title from user_feed uf join feed f on uf.feed_id=f.id where user_id=?"
	result := rssx.Find(stmt, []interface{}{userId}...)
	for i,v:=range result{
		 log.Info()
	}

}
