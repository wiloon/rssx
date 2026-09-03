package news

import (
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestMain(m *testing.M) {
	mr, err := miniredis.Run()
	if err != nil {
		panic("failed to start miniredis: " + err.Error())
	}
	// Must be set before the first redisx call (the pool initialises once).
	os.Setenv("REDIS_ADDRESS", mr.Addr())
	code := m.Run()
	mr.Close()
	os.Exit(code)
}

func TestNews_MarkReadThenIsRead(t *testing.T) {
	n := News{Id: "n1", FeedId: 7}

	if n.IsRead("0") {
		t.Fatal("article should not be read before MarkRead")
	}

	n.MarkRead(0)

	if !n.IsRead("0") {
		t.Fatal("article should be read after MarkRead")
	}
}

func TestNews_IsRead_ScopedPerFeed(t *testing.T) {
	a := News{Id: "shared", FeedId: 1}
	b := News{Id: "shared", FeedId: 2}

	a.MarkRead(0)

	if !a.IsRead("0") {
		t.Fatal("feed 1 article should be read")
	}
	if b.IsRead("0") {
		t.Fatal("same id under a different feed must not be read")
	}
}

func TestNews_SaveAndLoad(t *testing.T) {
	n := News{
		Id:          "save1",
		FeedId:      3,
		Title:       "Hello",
		Url:         "https://example.com/save1",
		Description: "<p>body</p>",
		PubDate:     "2026-01-01",
		Guid:        "save1",
		Score:       1234,
	}
	n.Save()

	got := News{Id: "save1"}
	got.Load()

	if got.Title != "Hello" || got.Url != "https://example.com/save1" {
		t.Fatalf("Load = %+v", got)
	}
	if got.Description != "<p>body</p>" {
		t.Fatalf("Load Description = %q", got.Description)
	}
	if got.Score != 1234 {
		t.Fatalf("Load Score = %d, want 1234", got.Score)
	}
}
