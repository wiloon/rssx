package feeds

import "rssx/feed"

// FeedRepository abstracts data access for feed subscriptions.
type FeedRepository interface {
	// FindByUserID returns all feeds subscribed by the given user.
	FindByUserID(userID string) ([]feed.Feed, error)
	// FindOrCreateByURL returns the existing feed for the URL, or creates it.
	FindOrCreateByURL(title, url string) (feed.Feed, error)
	// IsSubscribed reports whether the user is already subscribed to feedID.
	IsSubscribed(userID string, feedID int64) (bool, error)
	// Subscribe creates a subscription for the user to feedID.
	Subscribe(userID string, feedID int64) error
	// Unsubscribe removes the user's subscription. Returns false when not found.
	Unsubscribe(userID string, feedID int64) (bool, error)
}
