package config

import (
	"github.com/joho/godotenv" // Import godotenv
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"hrsale/models"
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
}

func InitializeDatabase() (*gorm.DB, error) {
	godotenv.Load(".env")

	dbConfig := DatabaseConfig{
		Host: os.Getenv("DB_HOST"),
	}
	portStr := os.Getenv("DB_PORT")
	dbConfig.Port, _ = strconv.Atoi(portStr)
	dbConfig.Username = os.Getenv("DB_USERNAME")
	dbConfig.Password = os.Getenv("DB_PASSWORD")
	dbConfig.DBName = os.Getenv("DB_NAME")

	dsn := dbConfig.Username + ":" + dbConfig.Password + "@tcp(" + dbConfig.Host + ":" + strconv.Itoa(dbConfig.Port) + ")/" + dbConfig.DBName + "?parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&models.Employee{})
	db.AutoMigrate(&models.Shift{})
	db.AutoMigrate(&models.Role{})
	db.AutoMigrate(&models.Admin{})
	db.AutoMigrate(&models.Department{})

	return db, nil
}
