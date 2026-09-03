package feeds

import (
	"errors"

	"rssx/feed"
)

// ErrURLConflict is returned by Update when the target URL already belongs to
// another feed (the feeds.url column is unique).
var ErrURLConflict = errors.New("feed url already exists")

// FeedRepository abstracts data access for feed subscriptions.
type FeedRepository interface {
	// FindByUserID returns all feeds subscribed by the given user.
	FindByUserID(userID string) ([]feed.Feed, error)
	// FindByID returns a single feed. The bool is false when no feed has that id.
	FindByID(feedID int64) (feed.Feed, bool, error)
	// FindOrCreateByURL returns the existing feed for the URL, or creates it.
	FindOrCreateByURL(title, url string) (feed.Feed, error)
	// Update changes a feed's title and URL. The bool is false when no feed has
	// that id; ErrURLConflict is returned when the URL is taken by another feed.
	Update(feedID int64, title, url string) (feed.Feed, bool, error)
	// Delete removes the feed and every subscription to it. Returns false when
	// no feed had that id.
	Delete(feedID int64) (bool, error)
	// Subscribers returns the user ids subscribed to feedID.
	Subscribers(feedID int64) ([]string, error)
	// IsSubscribed reports whether the user is already subscribed to feedID.
	IsSubscribed(userID string, feedID int64) (bool, error)
	// Subscribe creates a subscription for the user to feedID.
	Subscribe(userID string, feedID int64) error
	// Unsubscribe removes the user's subscription. Returns false when not found.
	Unsubscribe(userID string, feedID int64) (bool, error)
}
