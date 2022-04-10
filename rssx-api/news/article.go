package news

import (
	"fmt"
	"rssx/storage/redisx"
	log "rssx/utils/logger"
	"runtime"
	"strconv"
)

type Article interface {
	IsExistInStorage() bool
}

type DefaultArticle struct {
	FeedId int64
	Id     string
}

func (article *DefaultArticle) IsExistInStorage() bool {

	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panic: %v", p)
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Errorf("panic stack info: %s", fmt.Sprintf("%s", buf[:n]))
		}
	}()

	feedId := int(article.FeedId)
	articleId := article.Id
	var index int64
	result, err := redisx.GetConn().Do("ZRANK", feedNewsKey(feedId), articleId)
	log.Debugf("is exist in storage, feed id: %v, article id: %v, result: %v", feedId, articleId, result)
	if err != nil {
		log.Info(err.Error())
	}
	if result == nil {
		index = -1
	} else {
		index = result.(int64)
	}
	log.Debugf("find index by id: %v, index: %v", articleId, index)
	return index >= 0
}

const FeedNewsKeyPrefix string = "feed_news:"

func feedNewsKey(feedId int) string {
	key := FeedNewsKeyPrefix + strconv.Itoa(feedId)
	log.Debugf("get key of feed news: %v", key)
	return key
}
