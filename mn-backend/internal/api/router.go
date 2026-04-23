package router

import (
	"fmt"
	"net/http"
	"runtime"

	"moonick/internal/config"
	"moonick/internal/controller"
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

	// 静态文件服务
	r.Static("/public", "./public")

	fmt.Printf("Go Version %v\n", runtime.Version())

	appRepo := persistence.NewAppRepository()
	appService := service.NewAppService(appRepo)
	appController := controller.NewAppController(appService)

	metaController := controller.NewMetaController()

	v1 := r.Group("/api")

	v1.GET("/meta", metaController.GetMeta)

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
