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

func GetIndexByScore(key string, score int64) int64 {
	var rank int64
	if score == 0 {
		rank = 0
	} else {
		log.Debugf("get rank by score, key: %v, score: %v", key, score)
		r, err := GetConn().Do("ZRANGEBYSCORE", key, score, score)
		if err != nil {
			log.Error(err)
		}
		log.Infof("result: %v", r)

		foo := r.([]interface{})
		member := string(foo[0].([]byte))

		t, _ := GetConn().Do("ZRANK", key, member)
		rank = t.(int64)
	}
	log.Infof("rank: %v", rank)
	return rank
}
