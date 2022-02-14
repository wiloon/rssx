package rss

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"rssx/data"
	"rssx/feed"
	"rssx/feed/news/list"
	"rssx/news"
	"rssx/storage/redisx"
	"rssx/utils"
	"rssx/utils/config"
	log "rssx/utils/logger"
	"strconv"
	"strings"
	"time"
)

func Gc() {
	gcDuration, _ := time.ParseDuration(config.GetString("news.gc-duration", "24h"))
	ticker := time.NewTicker(gcDuration)
	for ; true; <-ticker.C {
		// clear cache
		//删除一段时间 之前 的数据。
		// 取一个月之前的score
		expireTime := config.GetString("news.expire-time", "-720h")
		d, _ := time.ParseDuration(expireTime)
		oneMonthAgo := time.Now().Add(d)
		oneMonthAgoMicroSecond := utils.TimeToMicroSecond(oneMonthAgo)

		tmp := data.FindUserFeeds(0)

		for _, v := range tmp {
			feedId := int(v.Id)
			feedNewsKey := list.FeedNewsKeyPrefix + strconv.Itoa(feedId)
			expiredNews := redisx.GetNewsIdListByScore(feedNewsKey, 0, oneMonthAgoMicroSecond)
			for _, newsId := range expiredNews {
				// 删除news
				redisx.DeleteNews(newsId)
			}
			//删除0 - score 的数据
			redisx.DeleteNewsIndex(feedNewsKey, 0, oneMonthAgoMicroSecond)
		}
		log.Info("clean cache done.")
	}
}

func Sync() {
	duration := time.Minute * time.Duration(config.GetInt("sync.duration"))
	ticker := time.NewTicker(duration)
	for ; true; <-ticker.C {
		log.Info("new sync start")
		//find all feeds
		feeds := data.FindFeeds()
		for _, feed := range feeds {
			SyncFeed(feed)
		}
		log.Info("sync tick done")
	}
}

func SyncFeed(feed feed.Feed) {
	log.Info("sync feed:", feed.Url)
	client := &http.Client{}
	request, err := http.NewRequest("GET", feed.Url, nil)
	if err != nil {
		log.Info("failed to sync feed:", feed)
		return
	}
	request.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.30 Safari/537.36")
	result, err := client.Do(request)

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

		log.Debugf("index:%v, title:%v, guid:%v", i, string(v.Title), guid)

		if strings.EqualFold(guid, "") {
			guid = v.Link
		}

		newsList := list.NewList(0, feed)

		// since duplicate pub date, and invalid pub date, set time.now() as score, make sure no duplicate score
		score := utils.TimeNowMicrosecond()

		newsId := utils.Md5(guid)
		if list.FindIndexById(int(feed.Id), newsId) == -1 {
			newsList.AppendNews(score, newsId)
			log.Debugf("score:%v, news id:%v", score, newsId)
			oneNews := news.News{
				Id:          newsId,
				FeedId:      feed.Id,
				Guid:        guid,
				Score:       score,
				Title:       v.Title,
				Description: v.Description,
				Url:         url,
				PubDate:     v.PubDate,
			}
			oneNews.Save()
		}

	}
}
