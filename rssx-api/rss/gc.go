package rss

import (
	"rssx/data"
	"rssx/feed/news/list"
	"rssx/storage/redisx"
	"rssx/utils"
	"rssx/utils/config"
	log "rssx/utils/logger"
	"strconv"
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
