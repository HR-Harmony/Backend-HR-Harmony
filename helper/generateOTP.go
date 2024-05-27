package helper

import (
	"math/rand"
	"strconv"
	"time"
)

// generateOTP menghasilkan OTP acak enam digit
func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	min := 100000
	max := 999999
	return strconv.Itoa(rand.Intn(max-min+1) + min)
}
