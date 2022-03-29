package list

import (
	"github.com/gin-gonic/gin"
	"rssx/feed"
	"rssx/news"
	"rssx/storage/redisx"
	"rssx/user"
	log "rssx/utils/logger"
	"strconv"
)

const FeedNewsKeyPrefix string = "feed_news:"
const PageSize int64 = 10

type NewsList struct {
	userId int
	feed   feed.Feed
}

func NewList(userId int, feed feed.Feed) *NewsList {
	var result = new(NewsList)
	result.userId = userId
	result.feed = feed
	return result
}

// AppendNews 新文章 ， 加入 到id集合
// score : 当前时间戳
func (newsList *NewsList) AppendNews(score int64, newsId string) {
	feedNewsKey := FeedNewsKeyPrefix + strconv.Itoa(int(newsList.feed.Id))
	_, _ = redisx.GetConn().Do("ZADD", feedNewsKey, score, newsId)
}

// FindNewsListByUserFeed 按用户和feed取一页未读文章
func FindNewsListByUserFeed(userId string, feedId int) []string {
	var newsList []string

	latestReadIndex := GetLatestReadIndex(userId, feedId)
	key := NewsListKey(feedId)
	unReadIndexStart := latestReadIndex + 1
	unReadIndexEnd := unReadIndexStart + PageSize - 1
	newsList = FindNewsListByRange(key, unReadIndexStart, unReadIndexEnd)
	log.Infof("find news list by feed, index start: %v, index enc: %v, list size: %v", unReadIndexStart, unReadIndexEnd, len(newsList))
	return newsList
}

func NewsListKey(feedId int) string {
	return FeedNewsKeyPrefix + strconv.Itoa(feedId)
}

// FindNewsListByRange 按索引取文章列表
func FindNewsListByRange(key string, start, end int64) []string {
	log.Debugf("find news list by rang, start: %v, end: %v", start, end)
	var newsidList []string

	result, err := redisx.GetConn().Do("ZRANGE", key, start, end)
	if err != nil {
		log.Info("failed to get news")
	}
	for _, v := range result.([]interface{}) {
		b := v.([]byte)
		newsId := string(b)
		log.Info("news id: " + newsId)
		newsidList = append(newsidList, newsId)
	}
	log.Debugf("find news list by rang, start: %v, end: %v, list size: %v", start, end, len(newsidList))
	return newsidList
}

// FinOneNewsByIndex 按索引取某一条文章的id
func FinOneNewsByIndex(index int64, feedId int) string {
	newsIdList := FindNewsListByRange(NewsListKey(feedId), index, index)
	if newsIdList != nil && len(newsIdList) > 0 {
		return newsIdList[0]
	}
	return ""
}

// FindNextId 找下一篇文章id
func FindNextId(feedId int, newsId string) string {
	var nextNewsId string
	index := FindIndexById(feedId, newsId)
	nextIndex := index + 1
	foo, _ := redisx.GetConn().Do("ZRANGE", feedNewsKey(feedId), nextIndex, nextIndex)
	if len(foo.([]interface{})) > 0 {
		nextNewsId = string(foo.([]interface{})[0].([]byte))

	} else {
		nextNewsId = ""
	}
	return nextNewsId
}

// FindPreviousNewsId 上一篇文章的id
func FindPreviousNewsId(feedId int, newsId string) string {
	var previousNewsId string
	index := FindIndexById(feedId, newsId)
	previousIndex := index - 1
	foo, _ := redisx.GetConn().Do("ZRANGE", feedNewsKey(feedId), previousIndex, previousIndex)
	if len(foo.([]interface{})) > 0 {
		previousNewsId = string(foo.([]interface{})[0].([]byte))

	} else {
		previousNewsId = ""
	}
	return previousNewsId
}

// feed_news:12
func feedNewsKey(feedId int) string {
	key := FeedNewsKeyPrefix + strconv.Itoa(feedId)
	log.Debugf("get key of feed news: %v", key)
	return key
}

// news list read index, value=sorted set range index, not score
const userFeedLatestReadIndex string = "read_index:"

// GetLatestReadIndex
//因为删除旧数据之后 索引值会变，所以用户 已读标记， 用score做为已读标记
//按score取index
//redis里保存 score, 取最新的未读索引时时先取score再用score取member,再用member取位置   -_-!!
func GetLatestReadIndex(userId string, feedId int) int64 {
	log.Debugf("get latest read index, user id: %v, feed id: %v", userId, feedId)
	score := 0
	latestReadIndexKey := userFeedLatestReadIndex + userId + ":" + strconv.Itoa(feedId)
	r, err := redisx.GetConn().Do("GET", latestReadIndexKey)
	if err != nil {
		log.Info(err.Error())
	}
	var rank int64
	if r != nil {
		b := r.([]byte)
		i := string(b)
		score, _ = strconv.Atoi(i) // score
		feedNewsKey := FeedNewsKeyPrefix + strconv.Itoa(feedId)
		rank = redisx.GetRankByScore(feedNewsKey, int64(score))
	} else {
		// 取不到score时
		rank = -1
	}

	log.Debugf("get latest read index, key: %v, score: %v, rank: %v", latestReadIndexKey, score, rank)
	return rank
}

// SetReadIndex 更新已读索引
// 存score值
func SetReadIndex(userId, feedId int, index int64) {
	log.Info("set read index, user id: %v, feed id: %v, index: %v", userId, feedId, index)
	// get score by rank
	feedNewsKey := FeedNewsKeyPrefix + strconv.Itoa(feedId)
	userFeedReadIndexKey := userFeedLatestReadIndex + strconv.Itoa(userId) + ":" + strconv.Itoa(feedId)
	score := redisx.GetScoreByRank(feedNewsKey, index)

	if score == 0 {
		log.Warn("invalid score, ignore")
		return
	}
	_, _ = redisx.GetConn().Do("SET", userFeedReadIndexKey, score)
	log.Debugf("set read index, score:%v", score)
}

// FindIndexById 按 article id 取索引
func FindIndexById(feedId int, newsId string) int64 {
	var index int64
	result, err := redisx.GetConn().Do("ZRANK", feedNewsKey(feedId), newsId)
	if err != nil {
		log.Info(err.Error())
	}
	if result == nil {
		index = -1
	} else {
		index = result.(int64)
	}
	log.Debugf("find index by id: %v, index: %v", newsId, index)
	return index
}

func Count(feedId int) int64 {
	var count int64
	result, err := redisx.GetConn().Do("ZCARD", feedNewsKey(feedId))
	if err != nil {
		log.Info(err.Error())
	}
	if result == nil {
		count = 0
	} else {
		count = result.(int64)
	}
	log.Debugf("feed: %v, news count: %v", feedId, count)
	return count
}

// LoadNewsListByFeed 按feed取一页
func LoadNewsListByFeed(feedId int) []news.News {
	var newsList []news.News
	if feedId == -1 {
		// find all news for all user feeds
		//	newsList = data.FindAllNewsForUser(user.DefaultId)
	} else {
		// by feed id
		newsIds := FindNewsListByUserFeed(user.DefaultId, feedId)
		for _, v := range newsIds {
			n := news.New(v)
			n.FeedId = int64(feedId)
			n.LoadTitle()
			n.LoadReadFlag(0)
			// calculate unread count

			newsList = append(newsList, *n)
			log.Debugf("append article: %v", n.Title)
		}
	}
	log.Debugf("new list size: %v", len(newsList))
	return newsList
}
func MarkWholePageAsRead(c *gin.Context) {

	feedId, _ := strconv.Atoi(c.Query("feedId"))
	readIndex := GetLatestReadIndex(user.DefaultId, feedId)
	// reset read index
	newIndex := readIndex + PageSize //新已读=旧值加每页数量
	count := Count(feedId)
	if newIndex >= count {
		newIndex = count - 1
	}
	log.Infof("mark page as read, feed id: %v,  last read index: %v, new index: %v, list count: %v",
		feedId, readIndex, newIndex, count)

	SetReadIndex(0, feedId, newIndex) //save
	// del read mark set,按feed删除
	news.DelReadMark(0, feedId)

	// load next page
	newsList := LoadNewsListByFeed(feedId)
	c.JSON(200, newsList)
}
func PreviousArticle(c *gin.Context) {
	currentNewsId := c.Query("newsId")
	feedId, _ := strconv.Atoi(c.Query("feedId"))
	log.Debugf(" load previous news feed id:%v, news id:%v", feedId, currentNewsId)
	index := FindIndexById(feedId, currentNewsId)
	newsIds := FindNewsListByRange(NewsListKey(feedId), index-1, index-1)
	previousNewsId := newsIds[0]
	previousNews := news.New(previousNewsId)
	previousNews.FeedId = int64(feedId)
	previousNews.Load()
	nextNewsId := FindNextId(feedId, previousNewsId)
	previousNews.NextId = nextNewsId
	c.JSON(200, previousNews)
}

// LoadArticles load one news
// 按 id 加载一篇文章
func LoadArticles(c *gin.Context) {
	feedId, _ := strconv.Atoi(c.Query("feedId"))
	newsId := c.Query("id")

	n := news.New(newsId)
	n.FeedId = int64(feedId)
	n.Load()
	log.Debugf("load one news, feed id:%v, news id:%v, title: %s", feedId, newsId, n.Title)

	nextNewsId := FindNextId(feedId, newsId)
	n.NextId = nextNewsId

	log.Info("show news:", n.Title, ", next id:", n.NextId)

	// 加载新的一条文章时要维护已读未读的边界 和 不连续的已读记录
	// 用户当前已读索引
	currentUserReadIndex := GetLatestReadIndex(user.DefaultId, feedId)
	// 当前文章的索引
	currentNewsIndex := FindIndexById(feedId, newsId)
	n.MarkRead(0)
	log.Debugf("currentUserReadIndex: %v, currentNewsIndex: %v", currentUserReadIndex, currentNewsIndex)

	nextUnReadIndex := findNextUserUnReadIndex(feedId, currentUserReadIndex)
	log.Debugf("currentUserReadIndex: %v, nextUnReadIndex: %v", currentUserReadIndex, nextUnReadIndex)
	if currentUserReadIndex == nextUnReadIndex {
		// 已读位置不连续，记录到已读集合
		n.MarkRead(0)
	} else {
		//已读文章是连续的，直接维护已读位置边界
		//更新用户已读索引
		SetReadIndex(0, feedId, nextUnReadIndex)
	}
	c.JSON(200, n)

}

/**
找到用户下一个未读索引
*/
func findNextUserUnReadIndex(feedId int, currentNewsIndex int64) int64 {
	log.Debugf("findNextUserUnReadIndex, feed id: %v, index: %v", feedId, currentNewsIndex)
	var result int64
	nextNewsIndex := currentNewsIndex + 1
	nextNewsId := FinOneNewsByIndex(nextNewsIndex, feedId)

	if nextNewsId == "" {
		result = currentNewsIndex
	} else {
		nextNews := news.New(nextNewsId)
		nextNews.FeedId = int64(feedId)
		if nextNews.IsRead(user.DefaultId) {
			result = findNextUserUnReadIndex(feedId, nextNewsIndex)
		} else {
			// 找到一条未读文章，退出
			result = currentNewsIndex
		}
	}

	log.Debugf("findNextUserUnReadIndex, feed id: %v, index: %v, result: %v", feedId, currentNewsIndex, result)
	return result
}
func LoadNewsList(c *gin.Context) {
	feedIdStr := c.Query("id")
	feedId, _ := strconv.Atoi(feedIdStr)
	log.Debugf("load news list by feed id: %v", feedId)

	newsList := LoadNewsListByFeed(feedId)

	c.JSON(200, newsList)

}
