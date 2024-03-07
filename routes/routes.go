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
	e.PUT("/shifts/:id", controllers.EditShiftByIDByAdmin(db, secretKey))
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

	//Project Admin
	e.POST("/projects", controllers.CreateProjectByAdmin(db, secretKey))
	e.GET("/projects", controllers.GetAllProjectsByAdmin(db, secretKey))
	e.GET("/projects/:id", controllers.GetProjectByIDByAdmin(db, secretKey))
	e.PUT("/projects/:id", controllers.UpdateProjectByIDByAdmin(db, secretKey))
	e.DELETE("/projects/:id", controllers.DeleteProjectByIDByAdmin(db, secretKey))

	//Task Admin
	e.POST("/tasks", controllers.CreateTaskByAdmin(db, secretKey))
	e.GET("/tasks", controllers.GetAllTasksByAdmin(db, secretKey))
	e.GET("/tasks/:id", controllers.GetTaskByIDByAdmin(db, secretKey))
	e.PUT("/tasks/:id", controllers.UpdateTaskByIDByAdmin(db, secretKey))
	e.DELETE("/tasks/:id", controllers.DeleteTaskByIDByAdmin(db, secretKey))

	//Case Admin
	e.POST("/cases", controllers.CreateCaseByAdmin(db, secretKey))
	e.GET("/cases", controllers.GetAllCasesByAdmin(db, secretKey))
	e.GET("/cases/:id", controllers.GetCaseByIDByAdmin(db, secretKey))
	e.PUT("/cases/:id", controllers.UpdateCaseByIDByAdmin(db, secretKey))
	e.DELETE("/cases/:id", controllers.DeleteCaseByIDByAdmin(db, secretKey))

	//Disciplinary Admin
	e.POST("/disciplinarys", controllers.CreateDisciplinaryByAdmin(db, secretKey))
	e.GET("/disciplinarys", controllers.GetAllDisciplinaryByAdmin(db, secretKey))
	e.GET("/disciplinarys/:id", controllers.GetDisciplinaryByIDByAdmin(db, secretKey))
	e.PUT("/disciplinarys/:id", controllers.UpdateDisciplinaryByIDByAdmin(db, secretKey))
	e.DELETE("/disciplinarys/:id", controllers.DeleteDisciplinaryByIDByAdmin(db, secretKey))

	//Helpdesk Admin
	e.POST("/helpdesks", controllers.CreateHelpdeskByAdmin(db, secretKey))
	e.GET("/helpdesks", controllers.GetAllHelpdeskByAdmin(db, secretKey))
	e.GET("/helpdesks/:id", controllers.GetHelpdeskByIDByAdmin(db, secretKey))
	e.PUT("/helpdesks/:id", controllers.UpdateHelpdeskByIDByAdmin(db, secretKey))
	e.DELETE("/helpdesks/:id", controllers.DeleteHelpdeskByIDByAdmin(db, secretKey))

	//Payroll
	e.GET("/payrolls", controllers.GetAllEmployeesPayrollInfo(db, secretKey))
	e.PUT("/payrolls/:payroll_id", controllers.UpdatePaidStatusByPayrollID(db, secretKey))
	e.GET("/payrolls/history", controllers.GetAllPayrollHistory(db, secretKey))

	//Finance
	e.POST("/finances", controllers.CreateFinanceByAdmin(db, secretKey))
	e.GET("/finances", controllers.GetAllFinanceByAdmin(db, secretKey))
	e.GET("/finances/:id", controllers.GetFinanceByIDByAdmin(db, secretKey))
	e.PUT("/finances/:id", controllers.UpdateFinanceByIDByAdmin(db, secretKey))
	e.DELETE("/finances/:id", controllers.DeleteFinanceByIDByAdmin(db, secretKey))

	//Deposit Category
	e.POST("/deposit_categories", controllers.CreateDepositCategoryByAdmin(db, secretKey))
	e.GET("/deposit_categories", controllers.GetAllDepositCategoriesByAdmin(db, secretKey))
	e.GET("/deposit_categories/:id", controllers.GetDepositCategoryByIDByAdmin(db, secretKey))
	e.PUT("/deposit_categories/:id", controllers.EditDepositCategoryByIDByAdmin(db, secretKey))
	e.DELETE("/deposit_categories/:id", controllers.DeleteDepositCategoryByIDByAdmin(db, secretKey))

	//Add Deposit
	e.POST("/add_deposits", controllers.AddDepositByAdmin(db, secretKey))
	e.GET("/add_deposits", controllers.GetAllAddDepositsByAdmin(db, secretKey))
	e.GET("/add_deposits/:id", controllers.GetDepositByIDByAdmin(db, secretKey))
	e.PUT("/add_deposits/:id", controllers.UpdateDepositByAdmin(db, secretKey))
	e.DELETE("/add_deposits/:id", controllers.DeleteDepositByIDByAdmin(db, secretKey))

	//Expense Category
	e.POST("/expense_categories", controllers.CreateExpenseCategoryByAdmin(db, secretKey))
	e.GET("/expense_categories", controllers.GetAllExpenseCategoriesByAdmin(db, secretKey))
	e.GET("/expense_categories/:id", controllers.GetExpenseCategoryByIDByAdmin(db, secretKey))
	e.PUT("/expense_categories/:id", controllers.EditExpenseCategoryByIDByAdmin(db, secretKey))
	e.DELETE("/expense_categories/:id", controllers.DeleteExpenseCategoryByIDByAdmin(db, secretKey))

	//Add Expense
	e.POST("/expenses", controllers.AddExpenseByAdmin(db, secretKey))
	e.GET("/expenses", controllers.GetAllAddExpensesByAdmin(db, secretKey))
	e.GET("/expenses/:id", controllers.GetExpenseByIDByAdmin(db, secretKey))
	e.PUT("/expenses/:id", controllers.UpdateExpenseByIDByAdmin(db, secretKey))
	e.DELETE("/expenses/:id", controllers.DeleteExpenseByIDByAdmin(db, secretKey))

	//Transaction
	e.GET("/transactions", controllers.GetAllTransactions(db, secretKey))

	//Attendance
	e.POST("/attendances", controllers.AddManualAttendanceByAdmin(db, secretKey))
	e.GET("/attendances", controllers.GetAllAttendanceByAdmin(db, secretKey))
	e.GET("/attendances/:id", controllers.GetAttendanceByIDByAdmin(db, secretKey))
	e.PUT("/attendances/:id", controllers.UpdateAttendanceByIDByAdmin(db, secretKey))
	e.DELETE("/attendances/:id", controllers.DeleteAttendanceByIDByAdmin(db, secretKey))

	//Exit Admin
	e.POST("/exits", controllers.CreateExitStatusByAdmin(db, secretKey))
	e.GET("/exits", controllers.GetAllExitStatusByAdmin(db, secretKey))
	e.GET("/exits/:id", controllers.GetExitStatusByIDByAdmin(db, secretKey))
	e.PUT("/exits/:id", controllers.UpdateExitStatusByAdmin(db, secretKey))
	e.DELETE("/exits/:id", controllers.DeleteExitStatusByIDByAdmin(db, secretKey))

	//Performance Goals
	e.POST("/goals_types", controllers.CreateGoalTypeByAdmin(db, secretKey))
	e.GET("/goals_types", controllers.GetAllGoalTypesByAdmin(db, secretKey))
	e.GET("/goals_types/:id", controllers.GetGoalTypeByIDByAdmin(db, secretKey))
	e.PUT("/goals_types/:id", controllers.UpdateGoalTypeByIDByAdmin(db, secretKey))
	e.DELETE("/goals_types/:id", controllers.DeleteGoalTypeByIDByAdmin(db, secretKey))

	//Performance Tracking Goals
	e.POST("/goals", controllers.CreateGoalByAdmin(db, secretKey))
	e.GET("/goals", controllers.GetAllGoalsByAdmin(db, secretKey))
	e.GET("/goals/:id", controllers.GetGoalByIDByAdmin(db, secretKey))
	e.PUT("/goals/:id", controllers.UpdateGoalByIDByAdmin(db, secretKey))
	e.DELETE("/goals/:id", controllers.DeleteGoalByIDByAdmin(db, secretKey))

	//Recruitment
	e.POST("/jobs", controllers.CreateNewJobByAdmin(db, secretKey))
	e.GET("/jobs", controllers.GetAllNewJobsByAdmin(db, secretKey))
	e.GET("/jobs/:id", controllers.GetNewJobByIDByAdmin(db, secretKey))
	e.PUT("/jobs/:id", controllers.UpdateNewJobByIDByAdmin(db, secretKey))
	e.DELETE("/jobs/:id", controllers.DeleteNewJobByIDByAdmin(db, secretKey))

	//Employee Admin
	e.POST("/admin/employees", controllers.CreateEmployeeAccountByAdmin(db, secretKey))
	e.GET("/admin/employees", controllers.GetAllEmployeesByAdmin(db, secretKey))
	e.GET("/admin/employees/:id", controllers.GetEmployeeByIDByAdmin(db, secretKey))
	e.PUT("/admin/employees/:id", controllers.UpdateEmployeeAccountByAdmin(db, secretKey))
	e.DELETE("/admin/employees/:id", controllers.DeleteEmployeeAccountByAdmin(db, secretKey))

	//Client Admin
	e.POST("/admin/clients", controllers.CreateClientAccountByAdmin(db, secretKey))
	e.GET("/admin/clients", controllers.GetAllClientsByAdmin(db, secretKey))
	e.GET("/admin/clients/:id", controllers.GetClientByIDByAdmin(db, secretKey))
	e.PUT("/admin/clients/:id", controllers.UpdateClientAccountByAdmin(db, secretKey))
	e.DELETE("/admin/clients/:id", controllers.DeleteClientAccountByAdmin(db, secretKey))

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
