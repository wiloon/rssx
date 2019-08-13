# rssx

### redis key
key: feed_news:feedId0
type: sort set
value: newsId
所有newsid， 按时间排序的

key: news:newsId0  
type: hash
value: 新闻内容， 数据量最大

key: read_index:userId0:feedId0
type: string
value: zset的score值
已读索引，整页加载某个feed时

key: read_mark:userId0:feedId0
type: set
value: newsId
已读集合， 已读新闻标记为灰色。

###
查询大于等于某一个score的第一条数据的索引
索引加每页显示条数确定当前页最后一个索引
按索引取数据
标记页 已读时