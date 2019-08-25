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

func GetIndexByScore(key string, score int64) {
	r, _ := GetConn().Do("ZRANGEBYSCORE", key, score, score)
	foo := r.([]interface{})
	member := string(foo[0].([]byte))

	rank, _ := GetConn().Do("ZRANK", key, member)
	log.Infof("rank: %v", rank)
}
