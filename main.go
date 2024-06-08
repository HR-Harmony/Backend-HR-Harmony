package main

import (
	"github.com/robfig/cron/v3"
	"hrsale/config"
	"hrsale/controllers"
	"log"
	"os"
)

func main() {
	router := config.SetupRouter()
	db, err := config.InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}

	c := cron.New()
	_, err = c.AddFunc("59 23 * * 1-5", func() {
		controllers.MarkAbsentEmployees(db)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add function to reset paid status every 25th of the month at 00:00
	_, err = c.AddFunc("0 0 25 * *", func() {
		controllers.ResetPaidStatus(db)
	})
	if err != nil {
		log.Fatal(err)
	}

	c.Start()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if PORT is not set
	}

	err = router.Start(":" + port)
	if err != nil {
		log.Fatal(err)
	}
}

/*
func main() {
	router := config.SetupRouter()
	db, err := config.InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}

	c := cron.New()
	_, err = c.AddFunc("59 23 * * 1-5", func() {
		controllers.MarkAbsentEmployees(db)
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add function to reset paid status every 25th of the month at 00:00
	_, err = c.AddFunc("0 0 25 * *", func() {
		controllers.ResetPaidStatus(db)
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
*/

/*
func main() {
	router := config.SetupRouter()
	err := router.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
*/
