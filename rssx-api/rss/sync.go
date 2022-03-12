package rss

import (
	"encoding/xml"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"io"
	"io/ioutil"
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
	p, _ := ants.NewPoolWithFunc(2, syncOneFeed)
	feedList := feeds.FindUserFeeds("0")
	for _, oneFeed := range *feedList {
		log.Debugf("invoke ant pool, feed id: %d", oneFeed.Id)
		err := p.Invoke(oneFeed)
		if err != nil {
			log.Error("failed to invoke feed sync")
			return
		}
	}
	p.Release()
}

func syncOneFeed(data interface{}) {
	oneFeed := data.(feed.Feed)
	log.Infof("sync feed, id: %d, url: %s", oneFeed.Id, oneFeed.Url)
	client := &http.Client{}
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

	var remoteFeedBody []byte
	if result.StatusCode == http.StatusOK {
		remoteFeedBody, _ = ioutil.ReadAll(result.Body)
		bodyString := string(remoteFeedBody) //todo if debug enabled convert to string
		log.Debug("get feed OK, feed body:", bodyString)
	}

	rss := Rss{}
	err = xml.Unmarshal(remoteFeedBody, &rss)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	for i, rssItem := range rss.Chan.Items {
		// compare and save
		url := rssItem.Link
		guid := rssItem.Guid

		log.Debugf("index:%v, title:%v, guid:%v", i, string(rssItem.Title), guid)

		if strings.EqualFold(guid, "") {
			guid = rssItem.Link
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
				Title:       rssItem.Title,
				Description: rssItem.Description,
				Url:         url,
				PubDate:     rssItem.PubDate,
			}
			oneNews.Save()
		}
	}
}
