package feeds

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"rssx/feed"
	zapLogger "rssx/utils/logger"
	"time"
)

var db *gorm.DB

func init() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)
	var err error
	db, err = gorm.Open(sqlite.Open("/data/rssx/rssx.db"), &gorm.Config{
		Logger: newLogger,
	})
	checkErr(err)
}
func FindUserFeeds(userId string) *[]feed.Feed {
	feeds := &[]feed.Feed{}
	db.Table("user_feeds").Select("feeds.id,feeds.title,feeds.url").Joins("join feeds on user_feeds.feed_id = feeds.id").Where("user_id = ?", userId).Order("user_feeds.sort desc").Find(feeds)
	return feeds
}

func checkErr(err error) {
	if err != nil {
		zapLogger.Errorf("err: %v", err)
	}
}
