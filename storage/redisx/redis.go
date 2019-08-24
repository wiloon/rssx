package redisx

import (
	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
	"github.com/wiloon/pingd-config"
)

var conn redis.Conn

func init() {
	connect()
}

func ZADD(key string, score int64, member string) {
	_, _ = GetConn().Do("ZADD", key, score, member)

}
func GetConn() redis.Conn {
	if conn == nil {
		conn = connect()
	}
	return conn
}
func connect() redis.Conn {
	var err error
	address := config.GetString("redis.address", "127.0.0.1:6379")
	conn, err = redis.Dial("tcp", address)
	if err != nil {
		log.Info("failed to connect to redis:" + err.Error())
	}
	log.Infof("connected to redis, address: %v, conn: %v", address, conn)
	return conn
}

//func GetLatestReadIndex(userId, feedId int) string {
//	result, _ := Conn.Do("GET", userFeedLatestReadIndex+strconv.Itoa(userId)+":"+strconv.Itoa(feedId))
//	return result.(string)
//}
//
//func SaveLatestReadIndex(userId, feedId int, score string) {
//	Conn.Do("SET", userFeedLatestReadIndex+strconv.Itoa(userId)+":"+strconv.Itoa(feedId), score)
//}

//
//func FindNewsByScore(feedId int, min, max string) {
//	Conn.Do("ZRANGEBYSCORE", feedNewsKeyPrefix+strconv.Itoa(feedId), "("+min, "("+max)
//}
