package routes

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/controllers"
	"hrsale/middleware"
	"io/ioutil"
	"net/http"
)

func ServeHTML(c echo.Context) error {
	htmlData, err := ioutil.ReadFile("index.html")
	if err != nil {
		return err
	}
	return c.HTML(http.StatusOK, string(htmlData))
}

func SetupRoutes(e *echo.Echo, db *gorm.DB) {
	e.Use(Logger())
	secretKey := []byte(middleware.GetSecretKeyFromEnv())
	e.GET("/", ServeHTML)

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

	//Designation Admin
	e.POST("/designations", controllers.CreateDesignationByAdmin(db, secretKey))
	e.GET("/designations", controllers.GetAllDesignationsByAdmin(db, secretKey))
	e.GET("/designations/:id", controllers.GetDesignationByID(db, secretKey))
	e.PUT("/designations/:id", controllers.UpdateDesignationByID(db, secretKey))
	e.DELETE("/designations/:id", controllers.DeleteDesignationByID(db, secretKey))

	//Policy Admin
	e.POST("/policies", controllers.CreatePolicyByAdmin(db, secretKey))
	e.GET("/policies", controllers.GetAllPoliciesByAdmin(db, secretKey))
	e.GET("/policies/:id", controllers.GetPolicyByIDByAdmin(db, secretKey))
	e.PUT("/policies/:id", controllers.UpdatePolicyByIDByAdmin(db, secretKey))
	e.DELETE("/policies/:id", controllers.DeletePolicyByIDByAdmin(db, secretKey))

	//Announcement Admin
	e.POST("/announcements", controllers.CreateAnnouncementByAdmin(db, secretKey))
	e.GET("/announcements", controllers.GetAnnouncementsByAdmin(db, secretKey))
	e.GET("/announcements/:id", controllers.GetAnnouncementByIDForAdmin(db, secretKey))
	e.PUT("/announcements/:id", controllers.UpdateAnnouncementForAdmin(db, secretKey))
	e.DELETE("/announcements/:id", controllers.DeleteAnnouncementForAdmin(db, secretKey))

	//Exit Admin
	e.POST("/exits", controllers.CreateExitStatusByAdmin(db, secretKey))
	e.GET("/exits", controllers.GetAllExitStatusByAdmin(db, secretKey))
	e.GET("/exits/:id", controllers.GetExitStatusByIDByAdmin(db, secretKey))
	e.PUT("/exits/:id", controllers.UpdateExitStatusByAdmin(db, secretKey))
	e.DELETE("/exits/:id", controllers.DeleteExitStatusByIDByAdmin(db, secretKey))

	//Employee Admin
	e.POST("/admin/employees", controllers.CreateEmployeeAccountByAdmin(db, secretKey))

	//Employee Exit Admin
	e.POST("/admin/employees/:id/exit", controllers.ExitEmployee(db, secretKey))
	e.GET("/admin/employees/exit", controllers.GetAllExitEmployees(db, secretKey))
	e.GET("/admin/employees/:id/exit", controllers.GetExitEmployeeByID(db, secretKey))
	e.DELETE("/admin/employees/:id/exit", controllers.DeleteExitEmployeeByID(db, secretKey))

	//Employee Login
	e.POST("/employee/signin", controllers.EmployeeLogin(db, secretKey))
	e.GET("/profile", controllers.EmployeeProfile(db, secretKey))

	// Chatbot untuk user dapat bertanya dengan Debot rekomendasi tempat wisata
	harmonyUsecase := controllers.NewHarmonyUsecase()
	e.POST("/chatbot", func(c echo.Context) error {
		return controllers.RecommendTraining(c, harmonyUsecase)
	})
}
