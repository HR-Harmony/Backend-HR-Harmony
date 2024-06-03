package helper

import (
	"strconv"
	"strings"
)

// ConvertHoursToMinutes mengonversi waktu dalam format jam (seperti "4.00 hours") menjadi total menit
func ConvertHoursToMinutes(hoursStr string) (int, error) {
	hoursStr = strings.TrimSuffix(hoursStr, " hours")
	hours, err := strconv.ParseFloat(hoursStr, 64)
	if err != nil {
		return 0, err
	}
	return int(hours * 60), nil
}
