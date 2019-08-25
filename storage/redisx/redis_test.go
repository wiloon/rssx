package redisx

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	config "github.com/wiloon/pingd-config"
	"os"
	"testing"
	"time"
)

func TestUUID(t *testing.T) {
	rootUUID, _ := uuid.FromString("5e4a8cfe-73df-4ca6-8089-18c189cc1aa3")

	for i := 0; i < 10; i++ {
		newsUUID := uuid.NewV5(rootUUID, "news_id")
		fmt.Println(newsUUID)
	}

}

func TestShar(t *testing.T) {
	s := "sha256 芳华"

	h := sha256.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	fmt.Printf("origin: %s, sha256 hash: %x\n", s, bs)

}

func TestMd5(t *testing.T) {
	s := "test str"
	data := []byte(s)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has) //将[]byte转成16进制

	fmt.Println(md5str1)

}

func Test0(t *testing.T) {
	log.Infof("start...")
	_ = os.Setenv("app_config", "/tmp/rssx-config-toml")
	config.LoadLocalConfig("rssx-config-toml")

	score0 := time.Now().UnixNano()
	ZADD("k0", score0, "news0")
	score1 := time.Now().UnixNano()
	ZADD("k0", score1, "news1")
	score2 := time.Now().UnixNano()
	ZADD("k0", score2, "news2")

	r, _ := GetConn().Do("ZRANGEBYSCORE", "k0", score1, score1)
	foo := r.([]interface{})
	s := string(foo[0].([]byte))
	fmt.Println(s)

	GetIndexByScore("k0", score1)
}
