package common

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestDedupeUserFeeds_LegacyTableMigrates reproduces a database created before
// the unique index existed: duplicates must be removed so AutoMigrate succeeds,
// and the index must then reject new duplicates.
func TestDedupeUserFeeds_LegacyTableMigrates(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	legacy := []string{
		`CREATE TABLE user_feeds (user_id text NOT NULL, feed_id integer NOT NULL, sort integer DEFAULT 0)`,
		`INSERT INTO user_feeds (user_id, feed_id) VALUES ('0', 1), ('0', 1), ('0', 1), ('0', 2), ('1', 1)`,
	}
	for _, stmt := range legacy {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("legacy setup %q: %v", stmt, err)
		}
	}

	if err := dedupeUserFeeds(db); err != nil {
		t.Fatalf("dedupeUserFeeds: %v", err)
	}
	if err := db.AutoMigrate(&UserFeed{}); err != nil {
		t.Fatalf("AutoMigrate after dedupe: %v", err)
	}

	var rows int64
	db.Model(&UserFeed{}).Count(&rows)
	if rows != 3 {
		t.Errorf("rows after dedupe = %d, want 3", rows)
	}
	if err := db.Create(&UserFeed{UserId: "0", FeedId: 1}).Error; err == nil {
		t.Error("inserting a duplicate subscription succeeded, want unique constraint error")
	}
}

func TestDedupeUserFeeds_NoTableIsNoop(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := dedupeUserFeeds(db); err != nil {
		t.Fatalf("dedupeUserFeeds on empty db: %v", err)
	}
}
