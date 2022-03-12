package feeds

import (
	"fmt"
	"testing"
	"time"
)

func Test0(t *testing.T) {
	list := FindUserFeeds("0")
	for _, v := range *list {
		fmt.Println(v)
	}
	time.Sleep(3 * time.Second)
}
