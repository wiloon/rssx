package rss

import (
	"bytes"
	"crypto/tls"
	"github.com/mmcdole/gofeed"
	"github.com/panjf2000/ants/v2"
	"io"
	"net/http"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/feeds"
	"rssx/news"
	"rssx/utils"
	"rssx/utils/config"
	log "rssx/utils/logger"
	"strings"
	"time"
)

func Sync() {
	duration := time.Minute * time.Duration(config.GetIntWithDefaultValue("sync.duration", 1))
	ticker := time.NewTicker(duration)
	for ; true; <-ticker.C {
		log.Info("new sync start")
		syncFeeds()
		log.Info("sync tick done")
	}
}

func syncFeeds() {
	p, err := ants.NewPoolWithFunc(2, syncOneFeed)
	if err != nil {
		log.Errorf("failed to create sync pool: %v", err)
		return
	}
	defer p.Release()
	feedList := feeds.FindSubscribedFeeds()
	log.Debugf("user feed list: %v", len(*feedList))
	for _, oneFeed := range *feedList {
		log.Debugf("invoke ant pool, feed id: %d", oneFeed.Id)
		err := p.Invoke(oneFeed)
		if err != nil {
			log.Errorf("failed to invoke feed sync for feed %d: %v", oneFeed.Id, err)
			return
		}
	}
}

// feedFetchTimeout bounds one feed download end to end; without it a server
// that never answers would pin a pool worker and stall every later sync tick.
const feedFetchTimeout = 30 * time.Second

// feedClient is used only for fetching feeds. Certificate verification is
// skipped because homelab feeds (e.g. rsshub.wiloon.lab) use a private CA the
// container does not trust; the setting stays scoped to this client.
var feedClient = newFeedClient()

func newFeedClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Transport: transport, Timeout: feedFetchTimeout}
}

func syncOneFeed(data interface{}) {
	oneFeed := data.(feed.Feed)
	log.Infof("sync feed, id: %d, url: %s", oneFeed.Id, oneFeed.Url)
	client := feedClient
	request, err := http.NewRequest("GET", oneFeed.Url, nil)
	if err != nil {
		log.Errorf("failed to sync feed: %v, err: %v", oneFeed, err)
		return
	}
	request.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.30 Safari/537.36")
	result, err := client.Do(request)

	if err != nil {
		log.Errorf("failed to sync feed: %v, err: %v", oneFeed, err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("failed to close body")
		}
	}(result.Body)

	if result.StatusCode != http.StatusOK {
		log.Errorf("failed to sync feed %d (%s): unexpected status %d", oneFeed.Id, oneFeed.Url, result.StatusCode)
		return
	}
	remoteFeedBody, err := io.ReadAll(result.Body)
	if err != nil {
		log.Errorf("failed to read feed %d (%s): %v", oneFeed.Id, oneFeed.Url, err)
		return
	}

	parsed, err := gofeed.NewParser().Parse(bytes.NewReader(remoteFeedBody))
	if err != nil {
		log.Errorf("failed to parse feed %d (%s): %v", oneFeed.Id, oneFeed.Url, err)
		return
	}

	cutoff := retentionCutoff()
	for i, item := range parsed.Items {
		guid := item.GUID
		log.Debugf("index:%v, title:%v, guid: %v", i, item.Title, guid)

		// GC drops articles older than the retention window from the index, so
		// the existence check below would treat them as new on the next sync.
		if published := itemTime(item); published != nil && published.Before(cutoff) {
			log.Debugf("skip expired item, feed id: %d, guid: %v, published: %v", oneFeed.Id, guid, published)
			continue
		}

		if guid == "" {
			guid = item.Link
		}
		if guid == "" {
			log.Debugf("skip item without guid or link, feed id: %d, title: %v", oneFeed.Id, item.Title)
			continue
		}

		newsList := list.NewList(0, oneFeed)

		// since duplicate pub date, and invalid pub date, set time.now() as score, make sure no duplicate score
		score := utils.TimeNowMicrosecond()

		newsId := utils.Md5(guid)
		// check if article is already exist in storage
		article := news.DefaultArticle{FeedId: oneFeed.Id, Id: newsId}
		if !article.IsExistInStorage() {
			newsList.AppendNews(score, newsId)
			log.Debugf("score:%v, news id:%v", score, newsId)
			oneNews := news.News{
				Id:          newsId,
				FeedId:      oneFeed.Id,
				Guid:        guid,
				Score:       score,
				Title:       item.Title,
				Description: itemBody(item),
				Url:         item.Link,
				PubDate:     itemDateText(item),
			}
			oneNews.Save()
		}
	}
}

// itemTime is when the item was published, falling back to its last update
// (Atom entries often carry only <updated>); nil if the feed gave neither.
func itemTime(item *gofeed.Item) *time.Time {
	if item.PublishedParsed != nil {
		return item.PublishedParsed
	}
	return item.UpdatedParsed
}

func itemDateText(item *gofeed.Item) string {
	if item.Published != "" {
		return item.Published
	}
	return item.Updated
}

// itemBody prefers the full content (content:encoded, Atom <content>) over the
// summary (<description>, Atom <summary>).
func itemBody(item *gofeed.Item) string {
	if strings.TrimSpace(item.Content) != "" {
		return item.Content
	}
	return item.Description
}
