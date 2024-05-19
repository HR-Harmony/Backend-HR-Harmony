package main

import (
	"github.com/robfig/cron/v3"
	"hrsale/config"
	"hrsale/models"
	"log"
	"time"

	"gorm.io/gorm"
)

func markAbsentEmployees(db *gorm.DB) {
	var employees []models.Employee
	db.Where("is_client = ? AND is_exit = ?", false, false).Find(&employees)

	today := time.Now().Format("2006-01-02")

	for _, employee := range employees {
		var existingAttendance models.Attendance
		result := db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error != nil {
			currentTime := time.Now()
			attendance := models.Attendance{
				EmployeeID:       employee.ID,
				Username:         employee.Username,
				FullNameEmployee: employee.FirstName + " " + employee.LastName,
				AttendanceDate:   today,
				Status:           "Absent",
				CreatedAt:        &currentTime,
			}
			db.Create(&attendance)
			log.Printf("Marked employee %s as absent on %s\n", employee.Username, today) // Add log here
		}
	}
}

func main() {
	router := config.SetupRouter()
	db, err := config.InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}

	c := cron.New()
	_, err = c.AddFunc("59 23 * * 1-5", func() {
		markAbsentEmployees(db)
	})
	if err != nil {
		log.Fatal(err)
	}

	c.Start()

	err = router.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

/*
func main() {
	router := config.SetupRouter()
	err := router.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
*/
