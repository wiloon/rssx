package feeds

import (
	"errors"
	"testing"

	"rssx/common"
)

func TestGormSubscribe_DuplicateReturnsErrAlreadySubscribed(t *testing.T) {
	common.InitForTesting()
	repo := NewGormFeedRepository(common.DB)

	f, err := repo.FindOrCreateByURL("Example", "https://example.com/dup")
	if err != nil {
		t.Fatalf("FindOrCreateByURL: %v", err)
	}
	if err := repo.Subscribe("u1", f.Id); err != nil {
		t.Fatalf("first Subscribe: %v", err)
	}
	if err := repo.Subscribe("u1", f.Id); !errors.Is(err, ErrAlreadySubscribed) {
		t.Fatalf("second Subscribe: err = %v, want ErrAlreadySubscribed", err)
	}
	if err := repo.Subscribe("u2", f.Id); err != nil {
		t.Fatalf("another user subscribing to the same feed: %v", err)
	}

	var rows int64
	common.DB.Model(&common.UserFeed{}).Where("feed_id = ?", f.Id).Count(&rows)
	if rows != 2 {
		t.Errorf("user_feeds rows for feed = %d, want 2", rows)
	}
}
