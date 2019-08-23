# rssx

## redis key
### 所有 newsid， 按时间排序的
key: feed_news:feedId0
type: sort set
value: newsId

### 新闻内容
key: news:newsId0  
type: hash
value: 新闻内容， 数据量最大

### 已读索引，整页加载某个feed时使用
key: read_index:userId0:feedId0
type: string
value: zset的score值

### 已读集合， 已读新闻标记为灰色。
key: read_mark:userId0:feedId0
type: set
value: newsId


###
- 显示feed 列表时，显示未读新闻数，feed总数-索引=未读数量
- 按feed id 加载一页未读新闻时，按索引range取
- 标记某一页为已读时，取上一次的已读索引位置， 加每页显示数，记录新的已读索引
- 加载某个feed的一页未读新闻时，查询大于等于某一个 score 的第一条数据的索引

### ZSCORE， 成员member的score值。
### ZRANGE, 返回指定区间内的成员
### ZRANGEBYSCORE, 返回有序集合中指定分数区间的成员列表 - 正序
### ZRANK, 返回指定成员的排名(位置值,0表示第一个成员) - 正序
### 移除有序集中，指定排名(rank)区间内的所有成员
### 移除有序集中，指定分数（score）区间内的所有成员。