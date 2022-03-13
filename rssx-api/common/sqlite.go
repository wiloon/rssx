package common

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	zapLog "rssx/utils/logger"
	"time"
)

var DB *gorm.DB

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
	DB, err = gorm.Open(sqlite.Open("/data/rssx/rssx.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		zapLog.Error("failed to init db")
	}
}
