package config

/*
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
	db.AutoMigrate(&models.AddExpense{})
	db.AutoMigrate(&models.NewJob{})
	db.AutoMigrate(&models.LeaveRequestType{})
	db.AutoMigrate(&models.LeaveRequest{})
	db.AutoMigrate(&models.OvertimeRequest{})
	db.AutoMigrate(&models.Trainer{})
	db.AutoMigrate(&models.TrainingSkill{})
	db.AutoMigrate(&models.Training{})
	db.AutoMigrate(&models.KPIIndicator{})
	db.AutoMigrate(&models.AdvanceSalary{})
	db.AutoMigrate(&models.RequestLoan{})
	db.AutoMigrate(&models.Note{})
	db.AutoMigrate(&models.KPAIndicator{})
	db.AutoMigrate(&models.ResetPasswordOTP{})
	db.AutoMigrate(&models.AdminResetPasswordOTP{})

	return db, nil
}
*/

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
		Host:     os.Getenv("DB_HOST"),
		Port:     func() int { port, _ := strconv.Atoi(os.Getenv("DB_PORT")); return port }(),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}

	dsn := "host=" + dbConfig.Host +
		" user=" + dbConfig.Username +
		" password=" + dbConfig.Password +
		" dbname=" + dbConfig.DBName +
		" port=" + strconv.Itoa(dbConfig.Port) +
		" sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.Employee{},
		&models.Shift{},
		&models.Role{},
		&models.Admin{},
		&models.Department{},
		&models.Exit{},
		&models.ExitEmployee{},
		&models.Designation{},
		&models.Policy{},
		&models.Announcement{},
		&models.Project{},
		&models.Task{},
		&models.Case{},
		&models.Disciplinary{},
		&models.Helpdesk{},
		&models.PayrollInfo{},
		&models.GoalType{},
		&models.Goal{},
		&models.Attendance{},
		&models.Finance{},
		&models.DepositCategory{},
		&models.AddDeposit{},
		&models.ExpenseCategory{},
		&models.AddExpense{},
		&models.NewJob{},
		&models.LeaveRequestType{},
		&models.LeaveRequest{},
		&models.OvertimeRequest{},
		&models.Trainer{},
		&models.TrainingSkill{},
		&models.Training{},
		&models.KPIIndicator{},
		&models.AdvanceSalary{},
		&models.RequestLoan{},
		&models.Note{},
		&models.KPAIndicator{},
		&models.ResetPasswordOTP{},
		&models.AdminResetPasswordOTP{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
