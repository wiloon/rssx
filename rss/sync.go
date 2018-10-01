package rss

import (
	"encoding/xml"
	"fmt"
	"github.com/wiloon/app-config"
	"github.com/wiloon/wiloon-log/log"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wiloon.com/rssx/data"
	"wiloon.com/rssx/feed"
	"wiloon.com/rssx/feed/news/list"
	"wiloon.com/rssx/news"
	"wiloon.com/rssx/utils"
)

func Sync() {
	syncNews := config.GetBool("syncNews")
	log.Debug("sync:" + strconv.FormatBool(syncNews))
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

		url := v.Link
		guid := v.Guid

		log.Infof("index:%v, title:%v, guid:%v", i, string(v.Title), guid)

		if strings.EqualFold(guid, "") {
			guid = v.Link
		}

		newsList := list.NewList(0, feed)

		// since duplicate pub date, and invalid pub date, set time.now() as score, make sure no duplicate score
		score := time.Now().UnixNano()

		newsId := utils.Md5(guid)
		newsList.AppendNews(score, newsId) //todo, check if exist
		log.Debugf("score:%v, news id:%v", score, newsId)
		oneNews := news.News{Id: newsId, FeedId: int64(feed.Id), Guid: guid, Score: score, Title: v.Title, Description: v.Description, Url: url, PubDate: v.PubDate}
		oneNews.Save()

	}
}
