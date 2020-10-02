package list

import (
	"fmt"
	config "github.com/wiloon/pingd-config"
	log "github.com/wiloon/pingd-log/logconfig/zaplog"
	"os"
	"rssx/storage/redisx"
	"testing"
	"time"
)

func TestReadIndex(t *testing.T) {
	fmt.Println(GetLatestReadIndex(0, 0))
}

func TestNewsExist(t *testing.T) {
	v := FindIndexById(0, "da660f185cc89a4a09e2578c65cdbc0")
	fmt.Println(v)
}

func TestCount(t *testing.T) {

	v := Count(0)
	fmt.Println(v)
}

func Test0(t *testing.T) {
	log.Infof("start...")
	_ = os.Setenv("app_config", "/tmp/rssx-config-toml")
	config.LoadLocalConfig("rssx-config-toml")

	score0 := time.Now().UnixNano()
	redisx.ZADD("k0", score0, "news0")
	score1 := time.Now().UnixNano()
	redisx.ZADD("k0", score1, "news1")
	score2 := time.Now().UnixNano()
	redisx.ZADD("k0", score2, "news2")

	r, _ := redisx.GetConn().Do("ZRANGEBYSCORE", "k0", score1, score1)
	foo := r.([]interface{})
	s := string(foo[0].([]byte))
	fmt.Println(s)
}
