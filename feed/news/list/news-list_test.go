package list

import (
	"fmt"
	"testing"
)

func TestReadIndex(t *testing.T) {
	fmt.Println(GetLatestReadIndex(0, 0))
}
