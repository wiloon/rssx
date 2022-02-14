package rss

import (
	"fmt"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	Sync()
}

func TestTime(t *testing.T) {
	fmt.Println(time.Now().UnixNano())
	fmt.Println(time.Now().UnixNano())
}
