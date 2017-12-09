package rss

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/xml"
	"wiloon.com/rssx/data"
	"github.com/wiloon/wiloon-log/log"
	"wiloon.com/rssx/feed"
	"time"
	"strings"
)

func Sync() {
	period := time.Duration(time.Second * 60)
	ticker := time.NewTicker(period)
	for ; true; <-ticker.C {
		//find all feeds
		feeds := data.FindFeeds()
		for _, feed := range feeds {
			SyncFeed(feed)
		}
	}
}

func SyncFeed(feed feed.Feed) {
	log.Info("sync feed:", feed)

	res, err := http.Get(feed.Url)
	if err != nil {
		log.Info("failed to sync feed:", feed)
		return
	}

	defer res.Body.Close()

	var remoteFeedBody []byte
	if res.StatusCode == http.StatusOK {
		remoteFeedBody, _ = ioutil.ReadAll(res.Body)
		bodyString := string(remoteFeedBody)

		log.Debug("get feed OK, feed body:", bodyString)
	}

	v := Rss{}
	err = xml.Unmarshal(remoteFeedBody, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	for i, v := range v.Chan.Items {
		// compare and save

		log.Info("index:", i, ", title:", string(v.Title))

		url := v.Link
		pubDate, _ := time.Parse(time.RFC1123Z, v.PubDate)
		guid := v.Guid
		if strings.EqualFold(guid, "") {
			guid = v.Link
		}
		//check if guid is exist
		found := data.FindNewsByGuid(guid)
		exist := len(found) == 1

		if !exist {
			data.SaveNews(feed.Id, v.Title, url, v.Description, pubDate, guid)
		}
	}
}
