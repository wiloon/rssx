package rss

import (
	"rssx/feed/news/list"
	"rssx/feeds"
	"rssx/storage/redisx"
	"rssx/utils"
	"rssx/utils/config"
	log "rssx/utils/logger"
	"strconv"
	"time"
)

// retentionCutoff returns the instant before which articles are expired, from
// news.expire-time (a negative duration, default -720h).
func retentionCutoff() time.Time {
	d, err := time.ParseDuration(config.GetString("news.expire-time", "-720h"))
	if err != nil {
		log.Errorf("invalid news.expire-time, using -720h: %v", err)
		d = -720 * time.Hour
	}
	return time.Now().Add(d)
}

func Gc() {
	gcDuration, _ := time.ParseDuration(config.GetString("news.gc-duration", "24h"))
	ticker := time.NewTicker(gcDuration)
	for ; true; <-ticker.C {
		// clear cache
		//删除一段时间 之前 的数据。
		oneMonthAgoMicroSecond := utils.TimeToMicroSecond(retentionCutoff())

		tmp := feeds.FindSubscribedFeeds()

		for _, v := range *tmp {
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
