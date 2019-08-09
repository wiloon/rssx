# rssx

### redis key

key: feed_news:feedId0
type: sort set
value: newsId
所有newsid

key: news:newsId0  
type: hash
value: 新闻内容

key: read_index:userId0:feedId0
type: string
value: zset的score值

key: read_mark:userId0:feedId0
type: set
value: newsId
已读集合

