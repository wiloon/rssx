package common

import (
	"log"
	"os"
	"path/filepath"
	"rssx/utils/config"
	zapLog "rssx/utils/logger"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// User 用户表
type User struct {
	Id         string `gorm:"primaryKey"`
	Name       string `gorm:"uniqueIndex;not null"`
	Password   string `gorm:"not null"`
	CreateTime string
}

// Feed 订阅源表
type Feed struct {
	Id    int64  `gorm:"primaryKey;autoIncrement"`
	Title string `gorm:"not null"`
	Url   string `gorm:"uniqueIndex;not null"`
}

// News 新闻表
type News struct {
	Id          string `gorm:"primaryKey"`
	FeedId      int64  `gorm:"index;not null"`
	Title       string
	Url         string `gorm:"uniqueIndex"`
	Description string `gorm:"type:text"`
	PubDate     string
	Guid        string `gorm:"uniqueIndex"`
	Score       int64
}

// UserFeed 用户订阅表
type UserFeed struct {
	UserId string `gorm:"index;not null"`
	FeedId int64  `gorm:"index;not null"`
	Sort   int    `gorm:"default:0"`
}

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

	// 从配置读取数据库路径，默认为 /var/lib/rssx-api/rssx-api.db
	rssxDb := config.GetString("sqlite.path", "/var/lib/rssx-api/rssx-api.db")

	// 确保数据库目录存在
	dbDir := filepath.Dir(rssxDb)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		zapLog.Error("failed to create db directory: %s, error: %v", dbDir, err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(rssxDb), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		zapLog.Error("failed to init db: %s, error: %v", rssxDb, err)
		return
	}

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(&User{}, &Feed{}, &News{}, &UserFeed{})
	if err != nil {
		zapLog.Error("failed to auto migrate tables, error: %v", err)
		return
	}

	zapLog.Info("database initialized successfully: %s", rssxDb)
}
