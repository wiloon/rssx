package utils

import (
	"fmt"
	"testing"
)

func TestStringToDate(t *testing.T) {
	str := "Fri, 18 Oct 2019 07:40:06 GMT"
	foo := StringToDateRFC1123(str)
	fmt.Println(foo)
}
