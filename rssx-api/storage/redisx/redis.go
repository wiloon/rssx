package redisx

import (
	"github.com/garyburd/redigo/redis"
	"github.com/wiloon/pingd-config"
	log "github.com/wiloon/pingd-log/logconfig/zaplog"
	"strconv"
)

var conn redis.Conn

func init() {
	// connect()
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

func GetRankByScore(key string, score int64) int64 {
	var rank int64
	if score == 0 {
		rank = 0
	} else {
		log.Debugf("get rank by score, key: %v, score: %v", key, score)
		r, err := GetConn().Do("ZRANGEBYSCORE", key, score, score)
		if err != nil {
			log.Error(err)
		}
		foo := r.([]interface{})
		if len(foo) == 0 {
			return 0
		}
		member := string(foo[0].([]byte))

		t, _ := GetConn().Do("ZRANK", key, member)
		rank = t.(int64)
	}
	log.Infof("got rank by score, score: %v, rank: %v", score, rank)
	return rank
}

func GetNewsIdListByScore(key string, scoreStart, scoreEnd int64) []string {
	var out []string
	r, err := GetConn().Do("ZRANGEBYSCORE", key, scoreStart, scoreEnd)
	if err != nil {
		log.Error(err)
	}

	if r != nil {
		foo := r.([]interface{})
		for _, v := range foo {
			member := string(v.([]byte))
			out = append(out, member)
		}
	}

	return out
}
func GetScoreByRank(key string, rank int64) int64 {
	log.Debugf("get score by rank, rank: %v", rank)
	result, err := GetConn().Do("ZRANGE", key, rank, rank)
	if err != nil {
		log.Info("failed to get news")
	}
	foo := result.([]interface{})
	var scoreInt int64
	if len(foo) > 0 {
		bar := foo[0].([]byte)
		member := string(bar)
		log.Debugf("rank: %v, member: %v", rank, member)
		t, _ := GetConn().Do("ZSCORE", key, member)
		score := t.([]byte)
		scoreStr := string(score)
		scoreInt, _ = strconv.ParseInt(scoreStr, 10, 64)
		log.Debugf("get score by rank, rank: %v, score: %v ", rank, scoreInt)
	}

	return scoreInt
}

func DeleteNews(newsId string) {
	_, _ = GetConn().Do("del", "news:"+newsId)

}

func DeleteNewsIndex(key string, min, max int64) {
	_, _ = GetConn().Do("ZREMRANGEBYSCORE", key, min, max)

}
