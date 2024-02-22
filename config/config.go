package config

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres" // Change the driver to PostgreSQL
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

	dsn := "user=" + dbConfig.Username + " password=" + dbConfig.Password + " dbname=" + dbConfig.DBName + " host=" + dbConfig.Host + " port=" + strconv.Itoa(dbConfig.Port) + " sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&models.Employee{})
	db.AutoMigrate(&models.Shift{})
	db.AutoMigrate(&models.Role{})
	db.AutoMigrate(&models.Admin{})
	db.AutoMigrate(&models.Department{})
	db.AutoMigrate(&models.Exit{})
	db.AutoMigrate(&models.ExitEmployee{})
	db.AutoMigrate(&models.Designation{})
	db.AutoMigrate(&models.Policy{})
	db.AutoMigrate(&models.Announcement{})
	db.AutoMigrate(&models.Project{})
	db.AutoMigrate(&models.Task{})
	db.AutoMigrate(&models.Case{})
	db.AutoMigrate(&models.Disciplinary{})
	db.AutoMigrate(&models.Helpdesk{})
	db.AutoMigrate(&models.PayrollInfo{})
	db.AutoMigrate(&models.GoalType{})
	db.AutoMigrate(&models.Goal{})
	db.AutoMigrate(&models.Attendance{})
	db.AutoMigrate(&models.Finance{})
	db.AutoMigrate(&models.DepositCategory{})
	db.AutoMigrate(&models.AddDeposit{})
	db.AutoMigrate(&models.ExpenseCategory{})

	return db, nil
}
