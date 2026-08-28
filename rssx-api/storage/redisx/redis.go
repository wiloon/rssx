package redisx

import (
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"

	"rssx/utils/config"
	log "rssx/utils/logger"
)

var (
	pool     *redis.Pool
	poolOnce sync.Once
)

func initPool() {
	address := config.GetString("redis.address", "127.0.0.1:6379")
	password := config.GetString("redis.password", "")

	pool = &redis.Pool{
		MaxIdle:         10,
		MaxActive:       50,
		IdleTimeout:     240 * time.Second,
		MaxConnLifetime: 30 * time.Minute,
		Wait:            true,
		Dial: func() (redis.Conn, error) {
			opts := []redis.DialOption{
				redis.DialConnectTimeout(3 * time.Second),
				redis.DialReadTimeout(3 * time.Second),
				redis.DialWriteTimeout(3 * time.Second),
			}
			if password != "" {
				opts = append(opts, redis.DialPassword(password))
			}
			conn, err := redis.Dial("tcp", address, opts...)
			if err != nil {
				log.Errorf("failed to connect to redis: %v", err)
				return nil, err
			}
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	log.Infof("redis pool initialized, address: %v", address)
}

// GetConn returns a connection from the pool.
// Callers MUST call Close() on it when done. For a pooled connection Close()
// returns it to the pool, it does not tear down the TCP connection.
func GetConn() redis.Conn {
	poolOnce.Do(initPool)
	return pool.Get()
}

func ZADD(key string, score int64, member string) {
	_, _ = Exec("ZADD", key, score, member)
}

// Exec runs a single command on a pooled connection and releases it afterwards.
func Exec(commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := GetConn()
	defer conn.Close()
	return conn.Do(commandName, args...)
}

// WithConn borrows one pooled connection for the duration of fn and releases it
// afterwards. Use it to pipeline several commands over a single round trip
// (conn.Send / conn.Flush / conn.Receive).
func WithConn(fn func(conn redis.Conn) error) error {
	conn := GetConn()
	defer conn.Close()
	return fn(conn)
}

// GetRankByScore returns the 0-based rank of the first member whose score equals
// the given score, i.e. the number of members ranked strictly before it.
// This is a single ZCOUNT call, equivalent to the old ZRANGEBYSCORE + ZRANK pair.
func GetRankByScore(key string, score int64) int64 {
	if score == 0 {
		return 0
	}
	rank, err := redis.Int64(Exec("ZCOUNT", key, "-inf", "("+strconv.FormatInt(score, 10)))
	if err != nil {
		log.Errorf("get rank by score failed, key: %v, score: %v, err: %v", key, score, err)
		return 0
	}
	log.Debugf("got rank by score, key: %v, score: %v, rank: %v", key, score, rank)
	return rank
}

func GetNewsIdListByScore(key string, scoreStart, scoreEnd int64) []string {
	var out []string
	r, err := Exec("ZRANGEBYSCORE", key, scoreStart, scoreEnd)
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
	result, err := Exec("ZRANGE", key, rank, rank)
	if err != nil {
		log.Errorf("failed to get news by rank: %v", err)
	}
	foo := result.([]interface{})
	var scoreInt int64
	if len(foo) > 0 {
		bar := foo[0].([]byte)
		member := string(bar)
		log.Debugf("rank: %v, member: %v", rank, member)
		t, _ := Exec("ZSCORE", key, member)
		score := t.([]byte)
		scoreStr := string(score)
		scoreInt, _ = strconv.ParseInt(scoreStr, 10, 64)
		log.Debugf("get score by rank, rank: %v, score: %v ", rank, scoreInt)
	}

	return scoreInt
}

func DeleteNews(newsId string) {
	_, _ = Exec("del", "news:"+newsId)
}

func DeleteNewsIndex(key string, min, max int64) {
	_, _ = Exec("ZREMRANGEBYSCORE", key, min, max)
}
