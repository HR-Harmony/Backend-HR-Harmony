package helper

import "time"

func convertToWIB(t time.Time) time.Time {
	wib, _ := time.LoadLocation("Asia/Jakarta")
	return t.In(wib)
}
