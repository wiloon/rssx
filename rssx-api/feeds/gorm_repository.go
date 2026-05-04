package feeds

import (
	"errors"

	"rssx/common"
	"rssx/feed"

	"gorm.io/gorm"
)

type gormFeedRepository struct {
	db *gorm.DB
}

// NewGormFeedRepository creates a FeedRepository backed by a GORM database.
func NewGormFeedRepository(db *gorm.DB) FeedRepository {
	return &gormFeedRepository{db: db}
}

func (r *gormFeedRepository) FindByUserID(userID string) ([]feed.Feed, error) {
	var result []feed.Feed
	err := r.db.Table("user_feeds").
		Select("feeds.id, feeds.title, feeds.url").
		Joins("join feeds on user_feeds.feed_id = feeds.id").
		Where("user_id = ?", userID).
		Order("user_feeds.sort desc").
		Find(&result).Error
	return result, err
}

func (r *gormFeedRepository) FindOrCreateByURL(title, url string) (feed.Feed, error) {
	f := common.Feed{Title: title, Url: url}
	if err := r.db.Where(common.Feed{Url: url}).FirstOrCreate(&f).Error; err != nil {
		return feed.Feed{}, err
	}
	return feed.Feed{Id: f.Id, Title: f.Title, Url: f.Url}, nil
}

func (r *gormFeedRepository) IsSubscribed(userID string, feedID int64) (bool, error) {
	var uf common.UserFeed
	err := r.db.Where("user_id = ? AND feed_id = ?", userID, feedID).First(&uf).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *gormFeedRepository) Subscribe(userID string, feedID int64) error {
	uf := common.UserFeed{UserId: userID, FeedId: feedID}
	return r.db.Create(&uf).Error
}

func (r *gormFeedRepository) Unsubscribe(userID string, feedID int64) (bool, error) {
	result := r.db.Where("user_id = ? AND feed_id = ?", userID, feedID).Delete(&common.UserFeed{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
