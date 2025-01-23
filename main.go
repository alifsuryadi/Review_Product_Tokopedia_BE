package main

import (
	"fmt"
	"os"
	"ulascan-be/config"
	"ulascan-be/constants"
	"ulascan-be/controller"
	"ulascan-be/database"
	_ "ulascan-be/docs"
	"ulascan-be/middleware"
	"ulascan-be/repository"
	"ulascan-be/routes"
	"ulascan-be/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title UlaScan BE API
// @version 1.2
// @description All provided API for Ulascan APP.
// @termsOfService http://swagger.io/terms/

// @contact.name Muhammad Hilman Al Ayubi
// @contact.email c010d4ky0983@bangkit.academy

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"
func main() {
	fmt.Println("STARTING...")

	var (
		// DATABASE
		db *gorm.DB = config.SetupDatabaseConnection()

		// REPOSITORY
		userRepository    repository.UserRepository    = repository.NewUserRepository(db)
		historyRepository repository.HistoryRepository = repository.NewHistoryRepository(db)

		// SERVICE
		jwtService       service.JWTService       = service.NewJWTService()
		userService      service.UserService      = service.NewUserService(userRepository, jwtService)
		historyService   service.HistoryService   = service.NewHistoryService(historyRepository)
		tokopediaService service.TokopediaService = service.NewTokopediaService()
		modelService     service.ModelService     = service.NewModelService()
		geminiService    service.GeminiService    = service.NewGeminiService()

		// CONTROLLER
		userController    controller.UserController    = controller.NewUserController(userService)
		historyController controller.HistoryController = controller.NewHistoryController(historyService)
		mlController      controller.MLController      = controller.NewMLController(tokopediaService, modelService, geminiService, historyService)
	)

	defer config.CloseDatabaseConnection(db)
	defer geminiService.CloseClient()

	fmt.Println("MIGRATING DATABASE...")
	if err := database.MigrateFresh(db); err != nil {
		panic(err)
	}
	fmt.Println("> Database Migrated")

	if os.Getenv("APP_ENV") == constants.ENUM_RUN_DEV {
		fmt.Println("RUNNING ON DEV ENV")
		fmt.Println("SEEDING DATABASE...")
		if err := database.Seeder(db); err != nil {
			panic(err)
		}
		fmt.Println("> Database Seeded")
	}

	// SERVER
	server := gin.Default()

	// Use middleware
	server.Use(middleware.Logger())
	server.Use(middleware.Recovery())
	server.Use(middleware.CORSMiddleware())

	// ROUTES
	apiGroup := server.Group("/api")
	routes.User(apiGroup, userController, jwtService)
	routes.ML(apiGroup, mlController, jwtService)
	routes.History(apiGroup, historyController, jwtService)

	// RUNING THE SERVER
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ip := os.Getenv("IP_INSTANCE")
	if ip == "" {
		port = "localhost:8080"
	}

	url := ginSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", ip))
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	if err := server.Run("0.0.0.0:" + port); err != nil {
		fmt.Println("Server failed to start: ", err)
		return
	}
}
