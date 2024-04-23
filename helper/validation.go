package helper

import (
	"regexp"
	"strconv"
	"time"
)

// IsValidPassword checks if the password meets the required criteria
func IsValidPassword(password string) bool {
	// Password must be at least 8 characters and contain a combination of letters and numbers
	return len(password) >= 8 && containsLetterAndNumber(password)
}

// IsValidPhoneNumber checks if the phone number contains only digits and has a minimum length of 10
func IsValidPhoneNumber(phoneNumber string) bool {
	// Phone number must contain only digits and have a minimum length of 10
	match, _ := regexp.MatchString("^[0-9]{10,}$", phoneNumber)
	return match
}

// containsLetterAndNumber checks if a string contains both letters and numbers
func containsLetterAndNumber(s string) bool {
	hasLetter := false
	hasNumber := false
	for _, char := range s {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		} else if char >= '0' && char <= '9' {
			hasNumber = true
		}
		if hasLetter && hasNumber {
			return true
		}
	}
	return false
}

func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func ContainsOnlyDigits(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// ParseStringToFloat mengonversi string menjadi float64 jika memungkinkan
func ParseStringToFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0 // Nilai default jika parsing gagal
	}
	return f
}

// ParseStringToInt parses a string to an integer.
func ParseStringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		// If conversion fails, return 0 (or handle the error according to your application's logic)
		return 0
	}
	return i
}

func IsValidAppraisalDateFormat(date string) bool {
	// Check if date has format mm-yy
	_, err := time.Parse("01-2006", date)
	return err == nil
}
