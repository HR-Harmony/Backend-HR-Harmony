package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"log"
	"os"
	"time"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenerateToken(username string, secretKey []byte) (string, error) {
	// Durasi token berlaku
	expirationTime := time.Now().Add(24 * time.Hour)

	// Membuat klaim JWT
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Membuat token JWT dengan metode HMAC
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Menandatangani token dengan kunci rahasia
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string, secretKey []byte) (string, error) {
	// Parsing token dengan secret key
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	// Memeriksa apakah token valid
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.Username, nil
	} else {
		return "", errors.New("Invalid token")
	}
}

func GenerateExpiredToken(expiredAt time.Time) (string, error) {
	// Membuat klaim JWT dengan waktu kadaluwarsa yang sudah lewat
	claims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredAt.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Membuat token JWT dengan metode None
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)

	// Menghasilkan token tanpa menandatanganinya
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetSecretKeyFromEnv() string {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SECRET_KEY tidak ditemukan di .env")
	}
	return secretKey
}
