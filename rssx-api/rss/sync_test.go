package rss

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/news"
	"rssx/utils"
)

var miniRedis *miniredis.Miniredis

func TestMain(m *testing.M) {
	mr, err := miniredis.Run()
	if err != nil {
		panic("failed to start miniredis: " + err.Error())
	}
	miniRedis = mr
	// Must be set before the first redisx call (the pool initialises once).
	os.Setenv("REDIS_ADDRESS", mr.Addr())
	code := m.Run()
	mr.Close()
	os.Exit(code)
}

func TestSyncOneFeed_SkipsItemsOlderThanRetention(t *testing.T) {
	miniRedis.FlushAll()
	recent := time.Now().Add(-time.Hour).UTC().Format(time.RFC1123Z)
	expired := time.Now().Add(-60 * 24 * time.Hour).UTC().Format(time.RFC1123Z)
	body := fmt.Sprintf(`<rss version="2.0"><channel><title>t</title>
<item><title>recent</title><link>https://example.com/recent</link><guid>recent</guid><pubDate>%s</pubDate></item>
<item><title>expired</title><link>https://example.com/expired</link><guid>expired</guid><pubDate>%s</pubDate></item>
<item><title>undated</title><link>https://example.com/undated</link><guid>undated</guid><pubDate>not a date</pubDate></item>
</channel></rss>`, recent, expired)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	withFeedClient(t, newFeedClient())

	syncOneFeed(feed.Feed{Id: 42, Url: srv.URL})

	if got := list.Count(42); got != 2 {
		t.Fatalf("indexed articles = %d, want 2 (recent + undated)", got)
	}
	if list.FindIndexById(42, utils.Md5("expired")) >= 0 {
		t.Error("expired item was indexed")
	}
	if miniRedis.Exists("news:" + utils.Md5("expired")) {
		t.Error("expired item was stored")
	}
	if list.FindIndexById(42, utils.Md5("recent")) < 0 {
		t.Error("recent item was not indexed")
	}
}

// syncBody serves body as feed feedID and syncs it once.
func syncBody(t *testing.T, feedID int64, body string) {
	t.Helper()
	miniRedis.FlushAll()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	withFeedClient(t, newFeedClient())
	syncOneFeed(feed.Feed{Id: feedID, Url: srv.URL})
}

func loadStored(t *testing.T, guid string) news.News {
	t.Helper()
	n := news.News{Id: utils.Md5(guid)}
	if err := n.Load(); err != nil {
		t.Fatalf("article %q was not stored: %v", guid, err)
	}
	return n
}

func TestSyncOneFeed_AtomFeed(t *testing.T) {
	updated := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	syncBody(t, 7, fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>atom</title>
  <entry>
    <id>tag:example.com,2026:entry-1</id>
    <title>Atom entry</title>
    <link rel="alternate" href="https://example.com/atom/1"/>
    <updated>%s</updated>
    <summary>short summary</summary>
    <content type="html">&lt;p&gt;full body&lt;/p&gt;</content>
  </entry>
</feed>`, updated))

	if got := list.Count(7); got != 1 {
		t.Fatalf("indexed articles = %d, want 1", got)
	}
	n := loadStored(t, "tag:example.com,2026:entry-1")
	if n.Title != "Atom entry" || n.Url != "https://example.com/atom/1" {
		t.Errorf("title/url = %q/%q", n.Title, n.Url)
	}
	if n.Description != "<p>full body</p>" {
		t.Errorf("description = %q, want the <content> body", n.Description)
	}
	if n.PubDate != updated {
		t.Errorf("pub date = %q, want <updated> %q", n.PubDate, updated)
	}
}

func TestSyncOneFeed_AtomEntryOnlyUpdatedIsRetentionChecked(t *testing.T) {
	expired := time.Now().Add(-60 * 24 * time.Hour).UTC().Format(time.RFC3339)
	syncBody(t, 8, fmt.Sprintf(`<feed xmlns="http://www.w3.org/2005/Atom"><title>a</title>
<entry><id>old</id><title>old</title><link href="https://example.com/old"/><updated>%s</updated></entry>
</feed>`, expired))

	if got := list.Count(8); got != 0 {
		t.Errorf("indexed articles = %d, want 0 for an entry past retention", got)
	}
}

func TestSyncOneFeed_RSSContentEncodedPreferredOverDescription(t *testing.T) {
	syncBody(t, 9, `<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/"><channel><title>t</title>
<item><title>full</title><link>https://example.com/full</link><guid>full</guid>
<description>teaser</description><content:encoded><![CDATA[<p>whole article</p>]]></content:encoded></item>
<item><title>teaser only</title><link>https://example.com/teaser</link>
<description>just a teaser</description></item>
</channel></rss>`)

	if got := loadStored(t, "full").Description; got != "<p>whole article</p>" {
		t.Errorf("description = %q, want content:encoded", got)
	}
	// Without a <guid> the link identifies the item.
	if got := loadStored(t, "https://example.com/teaser").Description; got != "just a teaser" {
		t.Errorf("description = %q, want <description>", got)
	}
}

// withFeedClient swaps feedClient for the duration of a test.
func withFeedClient(t *testing.T, c *http.Client) {
	t.Helper()
	orig := feedClient
	feedClient = c
	t.Cleanup(func() { feedClient = orig })
}

func TestNewFeedClient_HasTimeoutAndOwnTransport(t *testing.T) {
	c := newFeedClient()
	if c.Timeout != feedFetchTimeout {
		t.Errorf("timeout = %v, want %v", c.Timeout, feedFetchTimeout)
	}
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", c.Transport)
	}
	if tr == http.DefaultTransport {
		t.Error("feed client must not share http.DefaultTransport")
	}
	if dt := http.DefaultTransport.(*http.Transport); dt.TLSClientConfig != nil && dt.TLSClientConfig.InsecureSkipVerify {
		t.Error("http.DefaultTransport must keep certificate verification enabled")
	}
}

func TestSyncOneFeed_HangingServerTimesOut(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	defer srv.Close()
	defer close(release)

	c := newFeedClient()
	c.Timeout = 200 * time.Millisecond
	withFeedClient(t, c)

	done := make(chan struct{})
	go func() {
		syncOneFeed(feed.Feed{Id: 1, Url: srv.URL})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("syncOneFeed did not return for a server that never answers")
	}
}

func TestSyncOneFeed_ScoresByPublicationTime(t *testing.T) {
	miniRedis.FlushAll()
	older := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC1123Z)
	newer := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC1123Z)
	same := newer
	body := fmt.Sprintf(`<rss version="2.0"><channel><title>t</title>
<item><title>older</title><link>https://example.com/older</link><guid>older</guid><pubDate>%s</pubDate></item>
<item><title>newer</title><link>https://example.com/newer</link><guid>newer</guid><pubDate>%s</pubDate></item>
<item><title>same-a</title><link>https://example.com/same-a</link><guid>same-a</guid><pubDate>%s</pubDate></item>
<item><title>same-b</title><link>https://example.com/same-b</link><guid>same-b</guid><pubDate>%s</pubDate></item>
</channel></rss>`, older, newer, same, same)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	withFeedClient(t, newFeedClient())
	syncOneFeed(feed.Feed{Id: 11, Url: srv.URL})

	scoreOf := func(guid string) float64 {
		t.Helper()
		score, err := miniRedis.ZScore("feed_news:11", utils.Md5(guid))
		if err != nil {
			t.Fatalf("score of %s: %v", guid, err)
		}
		return score
	}
	if scoreOf("newer") <= scoreOf("older") {
		t.Errorf("newer score %v is not above older score %v", scoreOf("newer"), scoreOf("older"))
	}
	if scoreOf("same-a") == scoreOf("same-b") {
		t.Errorf("items that share a timestamp got the same score %v", scoreOf("same-a"))
	}
}

func TestSyncOneFeed_NonOKStatusReturns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnavailableForLegalReasons)
	}))
	defer srv.Close()
	withFeedClient(t, newFeedClient())

	syncOneFeed(feed.Feed{Id: 1, Url: srv.URL})
}
