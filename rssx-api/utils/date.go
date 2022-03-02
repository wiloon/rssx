package utils

import (
	"time"
)

func TimeNowMicrosecond() int64 {
	return TimeToMicroSecond(time.Now())
}

func TimeToMicroSecond(t time.Time) int64 {
	return t.UnixNano() / int64(time.Microsecond)
}

// StringToDateRFC1123 Sun, 06 Nov 1994 08:49:37 GMT; RFC 822, updated by RFC 1123
func StringToDateRFC1123(str string) time.Time {
	t, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", str)
	return t
}

func DateToStringYMDHMSZ(timestamp time.Time) string {
	return timestamp.Format("2006-01-02T15:04:05Z07:00")
}

func DateToStringYMD(timestamp time.Time) string {
	return timestamp.Format("2006-01-02")
}

func DateToStringYMDHM(timestamp time.Time) string {
	return timestamp.Format("2006-01-02 15:04")
}

func StringToDateYMDHMS(str string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", str)
	return t
}
