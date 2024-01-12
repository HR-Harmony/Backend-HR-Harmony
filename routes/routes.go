package routes

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/controllers"
	"hrsale/middleware"
)

func SetupRoutes(e *echo.Echo, db *gorm.DB) {
	e.Use(Logger())
	secretKey := []byte(middleware.GetSecretKeyFromEnv())
	e.POST("/admin/signup", controllers.RegisterAdminHR(db, secretKey))
	e.POST("/admin/signin", controllers.SignInAdmin(db, secretKey))
	e.GET("/verify", controllers.VerifyEmail(db))

	//Shift Admin
	e.POST("/shifts", controllers.CreateShiftByAdmin(db, secretKey))
	e.GET("/shifts", controllers.GetAllShiftsByAdmin(db, secretKey))
	e.GET("/shifts/:id", controllers.GetShiftByIDByAdmin(db, secretKey))
	e.PUT("/shifts/:id", controllers.EditShiftNameByIDByAdmin(db, secretKey))
	e.DELETE("/shifts/:id", controllers.DeleteShiftByIDByAdmin(db, secretKey))

	//Role Admin
	e.POST("/roles", controllers.CreateRoleByAdmin(db, secretKey))
	e.GET("/roles", controllers.GetAllRolesByAdmin(db, secretKey))
	e.GET("/roles/:id", controllers.GetRoleByIDByAdmin(db, secretKey))
	e.PUT("/roles/:id", controllers.EditRoleByIDByAdmin(db, secretKey))
	e.DELETE("/roles/:id", controllers.DeleteRoleByIDByAdmin(db, secretKey))

	//Department Admin
	e.POST("/departments", controllers.CreateDepartemntsByAdmin(db, secretKey))
	e.GET("/departments", controllers.GetAllDepartmentsByAdmin(db, secretKey))
	e.GET("/departments/:id", controllers.GetDepartmentByIDByAdmin(db, secretKey))
	e.PUT("/departments/:id", controllers.EditDepartmentByIDByAdmin(db, secretKey))
	e.DELETE("/departments/:id", controllers.DeleteDepartmentByIDByAdmin(db, secretKey))

	//Employee Admin
	e.POST("/admin/employees", controllers.CreateEmployeeAccountByAdmin(db, secretKey))

	//Employee Login
	e.POST("/employee/signin", controllers.EmployeeLogin(db, secretKey))
	e.GET("/profile", controllers.EmployeeProfile(db, secretKey))
}
