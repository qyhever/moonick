package router

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

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
	const (
		authIPRateLimitWindow = time.Minute
		authIPRateLimitLimit  = 10
	)

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
	db := initMySQLDB(cfg)
	userRepo := mysql.NewUserRepository(db)
	adminRepo := newAdminRepositoryFromConfig(cfg, db)
	registerCodeRepo := mysql.NewRegisterCodeRepository(db)
	r2Config := config.R2Config{}
	if cfg != nil {
		r2Config = cfg.R2
	}
	r2Storage := storage.NewR2(r2Config)
	fileService := service.NewFileService(r2Storage)
	authService := service.NewAuthService(userRepo, adminRepo, registerCodeRepo, jwtManager, service.NewPostalMailSender())
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
	attachController := controller.NewAttachController()
	tripController := controller.NewTripController(tripService, jwtManager)
	favoriteController := controller.NewFavoriteController(favoriteService, jwtManager)
	adminTripController := controller.NewAdminTripController(adminService)
	adminUserController := controller.NewAdminUserController(adminService)

	v1 := r.Group("/api")
	apiV1 := r.Group("/api/v1")
	adminV1 := r.Group("/api/admin/v1")

	v1.GET("/meta", metaController.GetMeta)
	apiV1.POST("/auth/register", middleware.NewIPRateLimit(authIPRateLimitWindow, authIPRateLimitLimit), authController.Register)
	apiV1.POST("/auth/code", middleware.NewIPRateLimit(authIPRateLimitWindow, authIPRateLimitLimit), authController.SendVerificationCode)
	apiV1.POST("/auth/password/reset", middleware.NewIPRateLimit(authIPRateLimitWindow, authIPRateLimitLimit), authController.ResetPassword)
	apiV1.POST("/auth/login", middleware.NewIPRateLimit(authIPRateLimitWindow, authIPRateLimitLimit), authController.Login)
	apiV1.POST("/auth/refresh", authController.Refresh)
	apiV1.GET("/auth/me", middleware.RequireUserAuth(jwtManager), authController.Me)
	apiV1.GET("/trips", tripController.List)
	apiV1.GET("/trips/:id", tripController.Detail)
	adminV1.POST("/auth/login", middleware.NewIPRateLimit(authIPRateLimitWindow, authIPRateLimitLimit), adminAuthController.Login)
	adminV1.POST("/auth/refresh", adminAuthController.Refresh)
	adminV1.GET("/auth/me", middleware.RequireAdminAuth(jwtManager), adminAuthController.Me)

	userGroup := apiV1.Group("/users")
	userGroup.Use(middleware.RequireUserAuth(jwtManager))
	{
		userGroup.GET("/me", userController.GetProfile)
		userGroup.PUT("/profile", userController.UpdateProfile)
		userGroup.PUT("/contact", userController.UpdateContact)
		userGroup.POST("/avatar", fileController.UploadAvatar)
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
	attach := v1.Group("/attach")
	{
		attach.POST("/add", attachController.Add)
		attach.POST("/upload", attachController.Upload)
		attach.DELETE("/delete", attachController.Delete)
		attach.GET("/list", attachController.List)
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

func initMySQLDB(cfg *config.Config) *sql.DB {
	if cfg == nil {
		return nil
	}

	dsn := config.BuildMySQLDSN(cfg)
	if strings.TrimSpace(dsn) == "" {
		return nil
	}

	db, err := mysql.OpenDB(dsn)
	if err != nil {
		panic(fmt.Errorf("初始化 MySQL 失败: %w", err))
	}

	mysql.SetDB(db)
	return db
}

func newAdminRepositoryFromConfig(cfg *config.Config, db *sql.DB) *mysql.AdminRepository {
	admin, err := buildAdminSeed(cfg)
	if err != nil {
		panic(fmt.Errorf("初始化管理员 seed 失败: %w", err))
	}
	repo := mysql.NewAdminRepositoryWithDB(db)
	if admin == nil {
		return repo
	}

	if db != nil {
		if err := repo.Upsert(context.Background(), *admin); err != nil {
			panic(fmt.Errorf("写入管理员 seed 失败: %w", err))
		}
		return repo
	}

	return mysql.NewAdminRepository(*admin)
}

func buildAdminSeed(cfg *config.Config) (*entity.Admin, error) {
	if cfg == nil {
		return nil, nil
	}

	adminCfg := cfg.Auth.Admin
	if strings.TrimSpace(adminCfg.Username) == "" || strings.TrimSpace(adminCfg.Password) == "" {
		return nil, nil
	}

	hash, err := password.Hash(adminCfg.Password)
	if err != nil {
		return nil, err
	}

	name := adminCfg.Name
	if strings.TrimSpace(name) == "" {
		name = adminCfg.Username
	}

	return &entity.Admin{
		ID:           1,
		Username:     adminCfg.Username,
		PasswordHash: hash,
		Name:         name,
		Status:       "active",
	}, nil
}
