# rssx

一个RSS阅读器

## redis key
### 某一个feed的 所有 news id， 按时间排序的
key: feed_news:feedId0
type: sort set， zset
value: newsId

### 文章内容
key: news:newsId0
type: hash
value: 文章内容， 数据量最大

## 记录用户阅读位置
用已读索引和已读集合记录用户阅读位置
已读索引 用于记录用户feed已读和未读的边界， 记录连续的已读未读位置
已读集合 用于记录已读边界外，用户分散阅读的新闻，记录不连续的已读集合

### 已读索引
key: read_index:userId0:feedId0
type: string
value: feed_news（zset）的score值

### 已读集合， 已读新闻标记为灰色
key: read_mark:userId0:feedId0
type: set
value: newsId

###
- 显示feed 列表时，显示未读新闻数，feed总数-索引=未读数量
- 按feed id 加载一页未读新闻时，按索引range取
- 标记某一页为已读时，取上一次的已读索引位置， 加每页显示数，记录新的已读索引
- 加载某个feed的一页未读新闻时，查询大于等于某一个 score 的第一条数据的索引

### ZSCORE， 成员member的score值
### ZRANGE, 返回指定区间内的成员
### ZRANGEBYSCORE, 返回有序集合中指定分数区间的成员列表 - 正序
### ZRANK, 返回指定成员的排名(位置值,0表示第一个成员) - 正序
### 移除有序集中，指定排名(rank)区间内的所有成员
### 移除有序集中，指定分数（score）区间内的所有成员

### 部署
#### mysql
```sql
CREATE DATABASE rssx DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;

DROP TABLE IF EXISTS feeds;
CREATE TABLE IF NOT EXISTS `feed` (
  `feed_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `title` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
  `url` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
  `deleted` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`feed_id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DELETE FROM feeds;
INSERT INTO feeds (id, `title`, `url`, `deleted`) VALUES
	(0, 'OS China', 'https://www.oschina.net/news/rss', 0),
	(1, 'InfoQ', 'http://www.infoq.com/cn/feed', 0),
	(3, 'CoolShell', 'http://coolshell.cn/feed', 0),
	(4, 'Solidot', 'http://feeds.feedburner.com/solidot', 0),
	(5, 'Autoblog', 'http://www.autoblog.com/rss.xml', 0),
	(7, 'Engadget', 'http://cn.engadget.com/rss.xml', 0),
	(8, 'Infozm', 'http://feed43.com/infzmnews.xml', 0),
	(9, 'huxiu', 'https://www.huxiu.com/rss/0.xml', 0),
	(10, '36ke', 'http://www.36kr.com/feed', 0),
	(11, 'FT', 'http://www.ftchinese.com/rss/news', 0);

DROP TABLE IF EXISTS `user`;
CREATE TABLE IF NOT EXISTS `user` (
  `user_id` bigint(20) DEFAULT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `create_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DELETE FROM `user`;
INSERT INTO `user` (`user_id`, `name`, `create_time`) VALUES
	(0, 'wiloon', '2017-12-07 22:10:49'),
	(1, 'foo', '2017-12-09 13:16:15');

DROP TABLE IF EXISTS user_feeds;
CREATE TABLE IF NOT EXISTS `user_feed` (
  `user_id` bigint(20) NOT NULL,
  `feed_id` bigint(20) NOT NULL,
  PRIMARY KEY (`user_id`,`feed_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DELETE FROM user_feeds;
INSERT INTO user_feeds (`user_id`, `feed_id`) VALUES
	(0, 0),
	(0, 1),
	(0, 2),
	(0, 3),
	(0, 4),
	(0, 5),
	(0, 6),
	(0, 7),
	(0, 8),
	(0, 9),
	(0, 10),
	(0, 11),
	(1, 0),
	(1, 1);

CREATE USER user0 IDENTIFIED BY 'password0';
grant all privileges on rssx.* to user0@'%' identified by 'password0';
```

### redis

### sqlite

```sql
CREATE TABLE if not exists users (  id char(36) PRIMARY KEY NOT NULL,  name varchar(50) DEFAULT NULL,  create_time timestamp DEFAULT NULL);

create table feed
(
    feed_id UNSIGNED BIG INT
        constraint feed_pk
            primary key,
    title   varchar(256),
    url     varchar(1024),
    deleted TINYINT
);


INSERT INTO `users` VALUES
                       (0,'wiloon','2017-12-07 22:10:49'),
                       (1,'foo','2017-12-09 13:16:15');

INSERT INTO feeds VALUES
                       (1,'InfoQ','https://www.infoq.cn/feed',0),
                       (3,'CoolShell','https://coolshell.cn/feed',0),
                       (4,'Solidot','https://www.solidot.org/index.rss',0),
                       (7,'Engadget-CN','https://chinese.engadget.com/rss.xml',0),
                       (8,'Infozm','https://node2.feed43.com/infzmnews.xml',0),
                       (9,'Engadget-EN','https://www.engadget.com/rss.xml',0),
                       (10,'36ke','https://www.36kr.com/feed',0),
                       (11,'FT','https://www.ftchinese.com/rss/news',0),
                       (12,'OS China','https://www.oschina.net/news/rss',0),
                       (13,'draveness','https://draveness.me/feed.xml',0);

INSERT INTO user_feeds VALUES
                            (0,0),
                            (0,1),
                            (0,2),
                            (0,3),
                            (0,4),
                            (0,5),
                            (0,6),
                            (0,7),
                            (0,8),
                            (0,9),
                            (0,10),
                            (0,11),
                            (0,12),
                            (0,13),
                            (1,0),
                            (1,1);

```