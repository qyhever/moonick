package router

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"moonick/internal/config"
	"moonick/internal/controller"
	"moonick/internal/middleware"
	"moonick/internal/model/entity"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/pkg/password"
	"moonick/internal/pkg/storage"
	"moonick/internal/repository/mysql"
	"moonick/internal/repository/persistence"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	isProd := config.IsProduction()
	// Gin 开启生产模式(默认是debug模式，会输出大量调试日志)
	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(middleware.RequestID())

	// 静态文件服务
	r.Static("/public", "./public")

	fmt.Printf("Go Version %v\n", runtime.Version())

	appRepo := persistence.NewAppRepository()
	appService := service.NewAppService(appRepo)
	appController := controller.NewAppController(appService)
	metaController := controller.NewMetaController()
	jwtManager := newJWTManager()
	cfg := config.GetConfig()
	userRepo := mysql.NewUserRepository()
	adminRepo := newAdminRepositoryFromConfig(cfg)
	r2Config := config.R2Config{}
	if cfg != nil {
		r2Config = cfg.R2
	}
	r2Storage := storage.NewR2(r2Config)
	fileService := service.NewFileService(r2Storage)
	authService := service.NewAuthService(userRepo, adminRepo, jwtManager)
	userService := service.NewUserService(userRepo, fileService)
	tripRepo := mysql.NewTripRepository()
	favoriteRepo := mysql.NewFavoriteRepository()
	tripService := service.NewTripService(tripRepo)
	favoriteService := service.NewFavoriteService(favoriteRepo, tripRepo)
	adminService := service.NewAdminService(userRepo, tripRepo, favoriteRepo)
	authController := controller.NewAuthController(authService, userService, jwtManager)
	userController := controller.NewUserController(userService, jwtManager)
	adminAuthController := controller.NewAdminAuthController(authService, jwtManager)
	fileController := controller.NewFileController(userService, jwtManager)
	tripController := controller.NewTripController(tripService, jwtManager)
	favoriteController := controller.NewFavoriteController(favoriteService, jwtManager)
	adminTripController := controller.NewAdminTripController(adminService)
	adminUserController := controller.NewAdminUserController(adminService)

	v1 := r.Group("/api")
	apiV1 := r.Group("/api/v1")
	adminV1 := r.Group("/api/admin/v1")

	v1.GET("/meta", metaController.GetMeta)
	apiV1.POST("/auth/register", authController.Register)
	apiV1.POST("/auth/login", authController.Login)
	apiV1.GET("/auth/me", middleware.RequireUserAuth(jwtManager), authController.Me)
	apiV1.GET("/trips", tripController.List)
	apiV1.GET("/trips/:id", tripController.Detail)
	adminV1.POST("/auth/login", adminAuthController.Login)
	adminV1.GET("/auth/me", middleware.RequireAdminAuth(jwtManager), adminAuthController.Me)

	userGroup := apiV1.Group("/users")
	userGroup.Use(middleware.RequireUserAuth(jwtManager))
	{
		userGroup.GET("/me", userController.GetProfile)
		userGroup.PUT("/profile", userController.UpdateProfile)
		userGroup.PUT("/contact", userController.UpdateContact)
	}

	tripGroup := apiV1.Group("/trips")
	tripGroup.Use(middleware.RequireUserAuth(jwtManager))
	{
		tripGroup.GET("/mine", tripController.ListMine)
		tripGroup.POST("", tripController.Create)
		tripGroup.PUT("/:id", tripController.Update)
		tripGroup.PATCH("/:id/status", tripController.UpdateStatus)
		tripGroup.POST("/:id/favorite", favoriteController.Toggle)
	}

	favoriteGroup := apiV1.Group("/favorites")
	favoriteGroup.Use(middleware.RequireUserAuth(jwtManager))
	{
		favoriteGroup.GET("", favoriteController.List)
		favoriteGroup.POST("/:tripId/toggle", favoriteController.Toggle)
	}

	meGroup := apiV1.Group("/me")
	meGroup.Use(middleware.RequireUserAuth(jwtManager))
	{
		meGroup.GET("/trips", tripController.ListMine)
		meGroup.GET("/favorites", favoriteController.List)
	}

	fileGroup := apiV1.Group("/files")
	fileGroup.Use(middleware.RequireUserAuth(jwtManager))
	{
		fileGroup.POST("/avatar", fileController.UploadAvatar)
	}

	adminProtectedGroup := adminV1.Group("")
	adminProtectedGroup.Use(middleware.RequireAdminAuth(jwtManager))
	{
		adminProtectedGroup.GET("/dashboard/summary", adminTripController.DashboardSummary)
		adminProtectedGroup.GET("/trips", adminTripController.List)
		adminProtectedGroup.GET("/trips/:id", adminTripController.Detail)
		adminProtectedGroup.PUT("/trips/:id", adminTripController.Update)
		adminProtectedGroup.GET("/users", adminUserController.List)
		adminProtectedGroup.GET("/users/:id", adminUserController.Detail)
		adminProtectedGroup.GET("/users/:id/trips", adminUserController.Trips)
	}

	app := v1.Group("/app")
	{
		app.POST("/getHelloInfo", appController.GetWebAuth)
	}
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "404",
		})
	})
	return r
}

func newJWTManager() *jwtpkg.Manager {
	cfg := config.GetConfig()
	if cfg == nil {
		return jwtpkg.NewManager(jwtpkg.Config{})
	}

	return jwtpkg.NewManager(jwtpkg.Config{
		Secret:          cfg.JWT.Secret,
		AccessTokenTTL:  cfg.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.JWT.RefreshTokenTTL,
	})
}

func newAdminRepositoryFromConfig(cfg *config.Config) *mysql.AdminRepository {
	if cfg == nil {
		return mysql.NewAdminRepository()
	}

	adminCfg := cfg.Auth.Admin
	if strings.TrimSpace(adminCfg.Username) == "" || strings.TrimSpace(adminCfg.Password) == "" {
		return mysql.NewAdminRepository()
	}

	hash, err := password.Hash(adminCfg.Password)
	if err != nil {
		return mysql.NewAdminRepository()
	}

	name := adminCfg.Name
	if strings.TrimSpace(name) == "" {
		name = adminCfg.Username
	}

	return mysql.NewAdminRepository(entity.Admin{
		ID:           1,
		Username:     adminCfg.Username,
		PasswordHash: hash,
		Name:         name,
		Status:       "active",
	})
}
