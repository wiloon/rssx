-- --------------------------------------------------------
-- Host:                         localhost
-- Server version:               10.1.29-MariaDB - MariaDB Server
-- Server OS:                    Linux
-- HeidiSQL Version:             9.4.0.5130
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;

-- Dumping data for table rssx.feed: ~11 rows (approximately)
DELETE FROM `feed`;
/*!40000 ALTER TABLE `feed` DISABLE KEYS */;
INSERT INTO `feed` (`feed_id`, `title`, `url`, `deleted`) VALUES
	(0, 'OS China', 'http://www.oschina.net/news/rss', 0),
	(1, 'InfoQ', 'http://www.infoq.com/cn/feed', 0),
	(2, '科学松鼠会', 'http://songshuhui.net/feed', 0),
	(3, 'CoolShell', 'http://coolshell.cn/feed', 0),
	(4, 'Solidot', 'http://feeds.feedburner.com/solidot', 0),
	(5, 'Autoblog', 'http://www.autoblog.com/rss.xml', 0),
	(6, 'Leica', 'http://www.leica.org.cn/feed.php', 0),
	(7, 'Engadget', 'http://cn.engadget.com/rss.xml', 0),
	(8, 'Infozm', 'http://feed43.com/infzmnews.xml', 0),
	(9, 'huxu', 'https://www.huxiu.com/rss/0.xml', 0),
	(10, '36ke', 'http://www.36kr.com/feed', 0),
	(11, 'FT', 'http://www.ftchinese.com/rss/news', 0);
/*!40000 ALTER TABLE `feed` ENABLE KEYS */;

-- Dumping data for table rssx.user: ~0 rows (approximately)
DELETE FROM `user`;
/*!40000 ALTER TABLE `user` DISABLE KEYS */;
INSERT INTO `user` (`user_id`, `name`, `create_time`) VALUES
	(0, 'wiloon', '2017-12-07 22:10:49'),
	(1, 'foo', '2017-12-09 13:16:15');
/*!40000 ALTER TABLE `user` ENABLE KEYS */;

-- Dumping data for table rssx.user_feed: ~13 rows (approximately)
DELETE FROM `user_feed`;
/*!40000 ALTER TABLE `user_feed` DISABLE KEYS */;
INSERT INTO `user_feed` (`user_id`, `feed_id`) VALUES
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
/*!40000 ALTER TABLE `user_feed` ENABLE KEYS */;

/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IF(@OLD_FOREIGN_KEY_CHECKS IS NULL, 1, @OLD_FOREIGN_KEY_CHECKS) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
