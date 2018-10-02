package redisx

import (
	"github.com/garyburd/redigo/redis"
	"github.com/wiloon/wiloon-log/log"
)

var Conn redis.Conn

func init() {
	var err error
	Conn, err = redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Info("failed to connect to redis:" + err.Error())
	}

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
