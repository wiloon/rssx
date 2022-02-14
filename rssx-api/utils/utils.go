package utils

import (
	"crypto/md5"
	"fmt"
	"time"
)

func Md5(str string) string {

	data := []byte(str)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has) //将[]byte转成16进制

	return md5str1
}

func TimeNowMicrosecond() int64 {
	return TimeToMicroSecond(time.Now())
}

func TimeToMicroSecond(t time.Time) int64 {
	return t.UnixNano() / int64(time.Microsecond)
}
