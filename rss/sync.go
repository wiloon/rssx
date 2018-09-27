package rss

import (
	"encoding/xml"
	"fmt"
	"github.com/wiloon/app-config"
	"github.com/wiloon/wiloon-log/log"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"wiloon.com/rssx/data"
	"wiloon.com/rssx/feed"
	"wiloon.com/rssx/storage/redisx"
)

func Sync() {
	syncNews := config.GetBool("syncNews")
	duration := time.Duration(time.Minute * time.Duration(config.GetInt("sync.duration")))
	ticker := time.NewTicker(duration)
	for ; syncNews; <-ticker.C {
		//find all feeds
		feeds := data.FindFeeds()
		for _, feed := range feeds {
			SyncFeed(feed)
		}
		log.Info("sync tick done")
	}
}

func SyncFeed(feed feed.Feed) {
	log.Info("sync feed:", feed)

	result, err := http.Get(feed.Url)
	if err != nil {
		log.Info("failed to sync feed:", feed)
		return
	}

	defer result.Body.Close()

	var remoteFeedBody []byte
	if result.StatusCode == http.StatusOK {
		remoteFeedBody, _ = ioutil.ReadAll(result.Body)
		bodyString := string(remoteFeedBody) //todo if debug enabled convert to string

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

		log.Info("index:", i, ", title:", string(v.Title)) //todo infof

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
			redisx.SaveNews(feed.Id, v.Title, url, v.Description, pubDate, guid)
		}
	}
}
