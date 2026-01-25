package common

import (
	"log"
	"os"
	"path/filepath"
	"rssx/utils/config"
	zapLog "rssx/utils/logger"
	"time"

	"gorm.io/driver/sqlite"
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

// getDatabasePath 获取数据库路径（支持环境变量）
func getDatabasePath() string {
	// 优先使用环境变量 DATABASE_PATH
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath != "" {
		return dbPath
	}

	// 其次使用配置文件，默认为 /data/rssx/rssx.db
	dbPath = config.GetString("sqlite.path", "/data/rssx/rssx.db")
	return dbPath
}

// seedData 插入默认数据
func seedData(db *gorm.DB) error {
	// 检查是否已有数据（避免重复插入）
	var feedCount int64
	if err := db.Model(&Feed{}).Count(&feedCount).Error; err != nil {
		return err
	}

	if feedCount > 0 {
		zapLog.Info("Database already has data, skipping seed")
		return nil
	}

	zapLog.Info("Seeding default data...")

	// 默认 RSS 源
	feeds := []Feed{
		{
			Url:   "https://hnrss.org/newest",
			Title: "Hacker News",
		},
		{
			Url:   "https://www.reddit.com/r/golang/.rss",
			Title: "r/golang",
		},
	}

	if err := db.Create(&feeds).Error; err != nil {
		return err
	}

	zapLog.Info("Seeded %d feeds", len(feeds))
	return nil
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

	// 获取数据库路径（支持环境变量）
	rssxDb := getDatabasePath()

	// 确保数据库目录存在
	dbDir := filepath.Dir(rssxDb)
	if dbDir != "." && dbDir != "/" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			zapLog.Error("failed to create db directory: %s, error: %v", dbDir, err)
		}
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

	// 初始化默认数据（seed data）
	if err := seedData(DB); err != nil {
		zapLog.Error("Warning: Failed to seed data: %v", err)
		// 不返回错误，允许应用继续运行
	}

	zapLog.Info("database initialized successfully: %s", rssxDb)
}
