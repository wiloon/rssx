package list

import (
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"

	"rssx/common"
	"rssx/feed"
	"rssx/news"
	"rssx/storage/redisx"
	"rssx/utils/jwt"
	log "rssx/utils/logger"
)

const FeedNewsKeyPrefix string = "feed_news:"
const PageSize int64 = 10

// AllFeedId is the pseudo-feed that aggregates every subscription.
const AllFeedId = -1

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
	_, _ = redisx.Exec("ZADD", feedNewsKey, score, newsId)
}

// ReserveScore returns preferred, or the next free score, so two articles in one
// feed never share a sort key. A shared score makes the read boundary ambiguous.
func ReserveScore(feedId int64, newsId string, preferred int64) int64 {
	if preferred < 1 {
		preferred = 1
	}
	key := NewsListKey(int(feedId))
	score := preferred
	for n := 0; n < 10000; n++ {
		members, err := redis.Strings(redisx.Exec("ZRANGEBYSCORE", key, score, score))
		if err != nil || !scoreTakenByOther(members, newsId) {
			return score
		}
		score++
	}
	return score
}

func scoreTakenByOther(members []string, newsId string) bool {
	for _, member := range members {
		if member != newsId {
			return true
		}
	}
	return false
}

// scoredArticle is one entry in a feed's index. Higher score means newer.
type scoredArticle struct {
	id     string
	score  int64
	feedId int
}

// FindNewsListByUserFeed returns one page of unread articles, newest first.
func FindNewsListByUserFeed(userId string, feedId int) []string {
	items := newestUnread(userId, feedId, int(PageSize))
	newsList := make([]string, len(items))
	for i, item := range items {
		newsList[i] = item.id
	}
	log.Debugf("find news list by feed, feed id: %v, list size: %v", feedId, len(newsList))
	return newsList
}

func NewsListKey(feedId int) string {
	return FeedNewsKeyPrefix + strconv.Itoa(feedId)
}

// PurgeFeed removes every trace of one feed from Redis: its article hashes, its
// article index, and the read marks / read boundary of the given subscribers.
// Used when a feed is deleted outright (not merely unsubscribed).
func PurgeFeed(feedId int, subscriberIds []string) {
	feedNewsKey := NewsListKey(feedId)
	for _, newsId := range FindNewsListByRange(feedNewsKey, 0, -1) {
		redisx.DeleteNews(newsId)
	}
	if _, err := redisx.Exec("DEL", feedNewsKey); err != nil {
		log.Errorf("purge feed, failed to drop index %v: %v", feedNewsKey, err)
	}
	for _, uid := range subscriberIds {
		ClearReadState(uid, feedId)
	}
}

// ClearReadState drops one user's read boundary and out-of-order read marks
// for one feed, e.g. when the user unsubscribes.
func ClearReadState(userId string, feedId int) {
	news.DelReadMark(userId, feedId)
	if _, err := redisx.Exec("DEL", ReadIndexKey(userId, feedId)); err != nil {
		log.Errorf("failed to drop read index for user %v, feed %v: %v", userId, feedId, err)
	}
}

// ReadIndexKey is the Redis key holding one user's read boundary (a score in
// feed_news:<feedId>) for one feed.
func ReadIndexKey(userId string, feedId int) string {
	return userFeedLatestReadIndex + userId + ":" + strconv.Itoa(feedId)
}

// FindNewsListByRange 按索引取文章列表
func FindNewsListByRange(key string, start, end int64) []string {
	log.Debugf("find news list by rang, start: %v, end: %v", start, end)
	var newsIdList []string

	result, err := redisx.Exec("ZRANGE", key, start, end)
	if err != nil {
		log.Errorf("failed to get news list by range, key: %v, err: %v", key, err)
		return newsIdList
	}
	for _, v := range result.([]interface{}) {
		b := v.([]byte)
		newsId := string(b)
		newsIdList = append(newsIdList, newsId)
	}
	log.Debugf("find news list by rang, start: %v, end: %v, list size: %v", start, end, len(newsIdList))
	return newsIdList
}

// FinOneNewsByIndex 按索引取某一条文章的id
func FinOneNewsByIndex(index int64, feedId int) string {
	newsIdList := FindNewsListByRange(NewsListKey(feedId), index, index)
	if newsIdList != nil && len(newsIdList) > 0 {
		return newsIdList[0]
	}
	return ""
}

// FindNextId returns the next newer article (higher score). previous-news uses
// it as the link back to the article the reader stepped away from.
func FindNextId(feedId int, newsId string) string {
	index := FindIndexById(feedId, newsId)
	if index < 0 {
		return ""
	}
	return FinOneNewsByIndex(index+1, feedId)
}

// olderID returns the next older article, which is the following row when the
// list is newest-first.
func olderID(feedId int, newsId string) string {
	index := FindIndexById(feedId, newsId)
	if index <= 0 {
		return ""
	}
	return FinOneNewsByIndex(index-1, feedId)
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
// 因为删除旧数据之后 索引值会变，所以用户 已读标记， 用score做为已读标记
// 按score取index
// redis里保存 score, 取最新的未读索引时时先取score再用score取member,再用member取位置   -_-!!
func GetLatestReadIndex(userId string, feedId int) int64 {
	latestReadIndexKey := ReadIndexKey(userId, feedId)
	r, err := redisx.Exec("GET", latestReadIndexKey)
	if err != nil {
		log.Errorf("get latest read index failed, key: %v, err: %v", latestReadIndexKey, err)
		return -1
	}
	if r == nil {
		// 没有已读标记时
		return -1
	}
	score, _ := strconv.Atoi(string(r.([]byte)))
	rank := redisx.GetRankByScore(NewsListKey(feedId), int64(score))
	log.Debugf("get latest read index, key: %v, score: %v, rank: %v", latestReadIndexKey, score, rank)
	return rank
}

// FeedUnreadCounts counts unread articles for each feed. An article is unread
// when its score is above the user's read boundary and its id is not in the
// per-user read set. The boundary is the legacy "read through this score"
// watermark; opening or marking a page records ids in the set so newer
// articles stay visible.
func FeedUnreadCounts(userId string, feedIds []int) map[int]int64 {
	counts := make(map[int]int64, len(feedIds))
	if len(feedIds) == 0 {
		return counts
	}

	type feedReadState struct {
		articles []scoredArticle
		boundary int64
		has      bool
		marks    map[string]bool
	}
	states := make([]feedReadState, len(feedIds))

	err := redisx.WithConn(func(conn redis.Conn) error {
		for _, fid := range feedIds {
			_ = conn.Send("ZREVRANGE", NewsListKey(fid), 0, -1, "WITHSCORES")
			_ = conn.Send("GET", ReadIndexKey(userId, fid))
			_ = conn.Send("SMEMBERS", news.ReadMarkKey(userId, int64(fid)))
		}
		if err := conn.Flush(); err != nil {
			return err
		}
		for i := range feedIds {
			reply, recvErr := conn.Receive()
			states[i].articles = scoredFromReply(reply, recvErr)
			raw, err := redis.Bytes(conn.Receive())
			if err == nil && len(raw) > 0 {
				if s, convErr := strconv.ParseInt(string(raw), 10, 64); convErr == nil {
					states[i].boundary = s
					states[i].has = true
				}
			}
			members, _ := redis.Strings(conn.Receive())
			states[i].marks = map[string]bool{}
			for _, id := range members {
				states[i].marks[id] = true
			}
		}
		return nil
	})
	if err != nil {
		log.Errorf("feed unread counts failed: %v", err)
		return counts
	}

	for i, fid := range feedIds {
		counts[fid] = countUnread(states[i].articles, states[i].boundary, states[i].has, states[i].marks)
	}
	return counts
}

func countUnread(articles []scoredArticle, boundary int64, hasBoundary bool, marks map[string]bool) int64 {
	var n int64
	for _, article := range articles {
		if hasBoundary && article.score <= boundary {
			break
		}
		if marks[article.id] {
			continue
		}
		n++
	}
	return n
}

func newestUnread(userId string, feedId, limit int) []scoredArticle {
	articles := feedArticlesNewestFirst(feedId)
	boundary, hasBoundary := readBoundary(userId, feedId)
	marks := readMarks(userId, feedId)
	out := make([]scoredArticle, 0, limit)
	for _, article := range articles {
		if hasBoundary && article.score <= boundary {
			break
		}
		if marks[article.id] {
			continue
		}
		article.feedId = feedId
		out = append(out, article)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func feedArticlesNewestFirst(feedId int) []scoredArticle {
	reply, err := redisx.Exec("ZREVRANGE", NewsListKey(feedId), 0, -1, "WITHSCORES")
	articles := scoredFromReply(reply, err)
	for i := range articles {
		articles[i].feedId = feedId
	}
	return articles
}

func scoredFromReply(reply interface{}, err error) []scoredArticle {
	if err != nil || reply == nil {
		return nil
	}
	parts, err := redis.Strings(reply, nil)
	if err != nil {
		return nil
	}
	out := make([]scoredArticle, 0, len(parts)/2)
	for i := 0; i+1 < len(parts); i += 2 {
		score, convErr := strconv.ParseInt(parts[i+1], 10, 64)
		if convErr != nil {
			continue
		}
		out = append(out, scoredArticle{id: parts[i], score: score})
	}
	return out
}

func readBoundary(userId string, feedId int) (int64, bool) {
	reply, err := redisx.Exec("GET", ReadIndexKey(userId, feedId))
	if err != nil || reply == nil {
		return 0, false
	}
	raw, ok := reply.([]byte)
	if !ok || len(raw) == 0 {
		return 0, false
	}
	score, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return 0, false
	}
	return score, true
}

func readMarks(userId string, feedId int) map[string]bool {
	marks := map[string]bool{}
	reply, err := redisx.Exec("SMEMBERS", news.ReadMarkKey(userId, int64(feedId)))
	if err != nil || reply == nil {
		return marks
	}
	ids, err := redis.Strings(reply, nil)
	if err != nil {
		return marks
	}
	for _, id := range ids {
		marks[id] = true
	}
	return marks
}

func userSubscribed(userId string, feedId int64) (bool, error) {
	var count int64
	err := common.DB.Model(&common.UserFeed{}).Where("user_id = ? AND feed_id = ?", userId, feedId).Count(&count).Error
	return count > 0, err
}

func subscribedFeedIDs(userId string) ([]int64, error) {
	var ids []int64
	err := common.DB.Model(&common.UserFeed{}).Where("user_id = ?", userId).Pluck("feed_id", &ids).Error
	return ids, err
}

// isNewer reports whether a should appear before b in a newest-first list.
func isNewer(a, b scoredArticle) bool {
	if a.score != b.score {
		return a.score > b.score
	}
	if a.feedId != b.feedId {
		return a.feedId < b.feedId
	}
	return a.id < b.id
}

func allViewPage(userId string, limit int) ([]scoredArticle, error) {
	feedIDs, err := subscribedFeedIDs(userId)
	if err != nil {
		return nil, err
	}
	merged := make([]scoredArticle, 0)
	for _, feedID := range feedIDs {
		merged = append(merged, newestUnread(userId, int(feedID), limit)...)
	}
	sort.SliceStable(merged, func(i, j int) bool {
		return isNewer(merged[i], merged[j])
	})
	if limit > 0 && len(merged) > limit {
		merged = merged[:limit]
	}
	return merged, nil
}

func articlesFromScored(userId string, items []scoredArticle) []news.News {
	if len(items) == 0 {
		return []news.News{}
	}
	byFeed := map[int][]string{}
	for _, item := range items {
		byFeed[item.feedId] = append(byFeed[item.feedId], item.id)
	}
	loaded := map[string]news.News{}
	for feedID, ids := range byFeed {
		for _, article := range news.LoadListForFeed(int64(feedID), userId, ids) {
			loaded[articleKey(feedID, article.Id)] = article
		}
	}
	out := make([]news.News, 0, len(items))
	for _, item := range items {
		article, ok := loaded[articleKey(item.feedId, item.id)]
		if !ok {
			continue
		}
		out = append(out, article)
	}
	return out
}

func articleKey(feedID int, id string) string {
	return strconv.Itoa(feedID) + ":" + id
}

// olderAcrossFeeds is the next older article across every feed the user
// subscribes to, compared with current.
func olderAcrossFeeds(userId string, current scoredArticle) (scoredArticle, bool) {
	feedIDs, err := subscribedFeedIDs(userId)
	if err != nil {
		log.Errorf("all view older article, load subscriptions: %v", err)
		return scoredArticle{}, false
	}
	var best scoredArticle
	found := false
	for _, feedID := range feedIDs {
		for _, article := range feedArticlesNewestFirst(int(feedID)) {
			article.feedId = int(feedID)
			if article.id == current.id && article.feedId == current.feedId {
				continue
			}
			if isNewer(article, current) || sameArticle(article, current) {
				continue
			}
			if !found || isNewer(article, best) {
				best = article
				found = true
			}
			break
		}
	}
	return best, found
}

func sameArticle(a, b scoredArticle) bool {
	return a.feedId == b.feedId && a.id == b.id
}

// SetReadIndex 更新已读索引
// 存score值
func SetReadIndex(userId string, feedId int, index int64) {
	log.Debugf("set read index, user id: %v, feed id: %v, index: %v", userId, feedId, index)
	// get score by rank
	feedNewsKey := FeedNewsKeyPrefix + strconv.Itoa(feedId)
	userFeedReadIndexKey := ReadIndexKey(userId, feedId)
	score := redisx.GetScoreByRank(feedNewsKey, index)

	if score == 0 {
		log.Warn("invalid score, ignore")
		return
	}
	_, _ = redisx.Exec("SET", userFeedReadIndexKey, score)
	log.Debugf("set read index, score:%v", score)
}

// FindIndexById 按 article id 取索引
func FindIndexById(feedId int, newsId string) int64 {
	var index int64
	result, err := redisx.Exec("ZRANK", feedNewsKey(feedId), newsId)
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
	result, err := redisx.Exec("ZCARD", feedNewsKey(feedId))
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

// LoadNewsListByFeed returns one page of unread articles, newest first.
// feedId -1 aggregates every feed the user subscribes to.
func LoadNewsListByFeed(userId string, feedId int) []news.News {
	var newsList []news.News
	if feedId == AllFeedId {
		page, err := allViewPage(userId, int(PageSize))
		if err != nil {
			log.Errorf("all view list failed, user %s: %v", userId, err)
			return []news.News{}
		}
		newsList = articlesFromScored(userId, page)
	} else {
		newsIds := FindNewsListByUserFeed(userId, feedId)
		newsList = news.LoadListForFeed(int64(feedId), userId, newsIds)
	}
	log.Debugf("new list size: %v", len(newsList))
	return newsList
}

func MarkWholePageAsRead(c *gin.Context) {
	userId := jwt.UserIdFromContext(c)
	feedId, ok := queryFeedId(c)
	if !ok {
		return
	}
	if !ensureFeedAccess(c, feedId) {
		return
	}
	markPageRead(userId, feedId)
	c.JSON(http.StatusOK, LoadNewsListByFeed(userId, feedId))
}

// markPageRead records the current newest-unread page as read. It does not
// move the score boundary: that watermark means "everything this old is read",
// and advancing it would hide articles the user has not seen yet.
func markPageRead(userId string, feedId int) {
	if feedId == AllFeedId {
		page, err := allViewPage(userId, int(PageSize))
		if err != nil {
			log.Errorf("mark all-view page read, user %s: %v", userId, err)
			return
		}
		for _, article := range page {
			(&news.News{Id: article.id, FeedId: int64(article.feedId)}).MarkRead(userId)
		}
		log.Infof("mark page as read, all view, user %s, count %d", userId, len(page))
		return
	}
	ids := FindNewsListByUserFeed(userId, feedId)
	for _, id := range ids {
		(&news.News{Id: id, FeedId: int64(feedId)}).MarkRead(userId)
	}
	log.Infof("mark page as read, feed id: %v, count: %v", feedId, len(ids))
}

// queryFeedId reads the feedId query param. It writes the 400 response itself
// and returns ok=false when the value is missing or not an integer.
func queryFeedId(c *gin.Context) (int, bool) {
	feedId, err := strconv.Atoi(c.Query("feedId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "feedId must be a valid integer"})
		return 0, false
	}
	return feedId, true
}

// loadNewsOrAbort loads one article, writing a 404 or 500 response and
// returning false when it cannot be loaded.
func loadNewsOrAbort(c *gin.Context, n *news.News) bool {
	if err := n.Load(); err != nil {
		if errors.Is(err, news.ErrNewsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return false
		}
		log.Errorf("failed to load news %v: %v", n.Id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storage error"})
		return false
	}
	return true
}

func PreviousArticle(c *gin.Context) {
	currentNewsId := c.Query("newsId")
	feedId, ok := queryFeedId(c)
	if !ok {
		return
	}
	if !ensureFeedAccess(c, feedId) {
		return
	}
	if feedId == AllFeedId {
		previousAcrossFeeds(c, currentNewsId)
		return
	}
	log.Debugf(" load previous news feed id:%v, news id:%v", feedId, currentNewsId)
	index := FindIndexById(feedId, currentNewsId)
	if index < 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	if index == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no previous article"})
		return
	}
	previousNewsId := FinOneNewsByIndex(index-1, feedId)
	if previousNewsId == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no previous article"})
		return
	}
	previousNews := news.New(previousNewsId)
	previousNews.FeedId = int64(feedId)
	if !loadNewsOrAbort(c, previousNews) {
		return
	}
	previousNews.NextId = FindNextId(feedId, previousNewsId)
	c.JSON(http.StatusOK, previousNews)
}

// LoadArticles load one news
// 按 id 加载一篇文章
func LoadArticles(c *gin.Context) {
	feedId, ok := queryFeedId(c)
	if !ok {
		return
	}
	if !ensureFeedAccess(c, feedId) {
		return
	}
	newsId := c.Query("id")
	if newsId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	n := news.New(newsId)
	if !loadNewsOrAbort(c, n) {
		return
	}
	if feedId != AllFeedId && n.FeedId != int64(feedId) {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	if feedId == AllFeedId && !ensureFeedAccess(c, int(n.FeedId)) {
		return
	}
	log.Debugf("load one news, feed id:%v, news id:%v, title: %s", n.FeedId, newsId, n.Title)

	userId := jwt.UserIdFromContext(c)
	if feedId == AllFeedId {
		if older, ok := olderAcrossFeeds(userId, scoredArticle{id: n.Id, score: n.Score, feedId: int(n.FeedId)}); ok {
			n.NextId = older.id
		}
	} else {
		n.NextId = olderID(feedId, newsId)
	}
	n.MarkRead(userId)
	log.Info("show news:", n.Title, ", next id:", n.NextId)
	c.JSON(http.StatusOK, n)
}

func LoadNewsList(c *gin.Context) {
	feedId, ok := queryListFeedID(c)
	if !ok {
		return
	}
	if !ensureFeedAccess(c, feedId) {
		return
	}
	log.Debugf("load news list by feed id: %v", feedId)
	c.JSON(http.StatusOK, LoadNewsListByFeed(jwt.UserIdFromContext(c), feedId))
}

// queryListFeedID reads the id query param used by GET /news-list.
func queryListFeedID(c *gin.Context) (int, bool) {
	raw := c.Query("id")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
		return 0, false
	}
	feedId, err := strconv.Atoi(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
		return 0, false
	}
	return feedId, true
}

// ensureFeedAccess allows the all-view pseudo-feed and rejects feeds the
// caller does not subscribe to. Missing subscriptions are 404 so feed ids
// cannot be probed for someone else's articles.
func ensureFeedAccess(c *gin.Context, feedId int) bool {
	if feedId == AllFeedId {
		return true
	}
	if feedId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "feedId must be a valid integer"})
		return false
	}
	ok, err := userSubscribed(jwt.UserIdFromContext(c), int64(feedId))
	if err != nil {
		log.Errorf("subscription check failed, feed %d: %v", feedId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return false
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "feed not found"})
		return false
	}
	return true
}

func previousAcrossFeeds(c *gin.Context, currentNewsId string) {
	if currentNewsId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "newsId is required"})
		return
	}
	current := news.New(currentNewsId)
	if !loadNewsOrAbort(c, current) {
		return
	}
	if !ensureFeedAccess(c, int(current.FeedId)) {
		return
	}
	userId := jwt.UserIdFromContext(c)
	older, ok := olderAcrossFeeds(userId, scoredArticle{id: current.Id, score: current.Score, feedId: int(current.FeedId)})
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "no previous article"})
		return
	}
	previous := news.New(older.id)
	if !loadNewsOrAbort(c, previous) {
		return
	}
	previous.NextId = currentNewsId
	c.JSON(http.StatusOK, previous)
}
