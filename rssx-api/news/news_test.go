package news

import (
	"fmt"
	"testing"
)

func TestNews_MarkRead(t *testing.T) {
	n := News{Id: "d033b8dfe091ef0262bd54dcb49bc04e", FeedId: 0}
	n.MarkRead(0)
}
func TestNews_IsRead(t *testing.T) {
	n := News{Id: "d033b8dfe091ef0262bd54dcb49bc04e", FeedId: 0}
	fmt.Println(n.IsRead(0))
}
