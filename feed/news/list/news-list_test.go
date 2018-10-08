package list

import (
	"fmt"
	"testing"
)

func TestReadIndex(t *testing.T) {
	fmt.Println(GetLatestReadIndex(0, 0))
}

func TestNewsExist(t *testing.T) {
	v:=FindIndexById(0,"da660f185cc89a4a09e2578c65cdbc0")
	fmt.Println(v)
}