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

func (r *gormFeedRepository) FindByID(feedID int64) (feed.Feed, bool, error) {
	var f common.Feed
	err := r.db.First(&f, feedID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return feed.Feed{}, false, nil
		}
		return feed.Feed{}, false, err
	}
	return feed.Feed{Id: f.Id, Title: f.Title, Url: f.Url}, true, nil
}

func (r *gormFeedRepository) Update(feedID int64, title, url string) (feed.Feed, bool, error) {
	var f common.Feed
	if err := r.db.First(&f, feedID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return feed.Feed{}, false, nil
		}
		return feed.Feed{}, false, err
	}

	var clash common.Feed
	err := r.db.Where("url = ? AND id <> ?", url, feedID).First(&clash).Error
	if err == nil {
		return feed.Feed{}, true, ErrURLConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return feed.Feed{}, true, err
	}

	f.Title = title
	f.Url = url
	if err := r.db.Model(&common.Feed{}).Where("id = ?", feedID).
		Updates(map[string]interface{}{"title": title, "url": url}).Error; err != nil {
		return feed.Feed{}, true, err
	}
	return feed.Feed{Id: f.Id, Title: title, Url: url}, true, nil
}

func (r *gormFeedRepository) Delete(feedID int64) (bool, error) {
	var removed bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("feed_id = ?", feedID).Delete(&common.UserFeed{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&common.Feed{}, feedID)
		if result.Error != nil {
			return result.Error
		}
		removed = result.RowsAffected > 0
		return nil
	})
	return removed, err
}

func (r *gormFeedRepository) Subscribers(feedID int64) ([]string, error) {
	var ids []string
	err := r.db.Model(&common.UserFeed{}).
		Where("feed_id = ?", feedID).
		Pluck("user_id", &ids).Error
	return ids, err
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
