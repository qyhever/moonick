package controller

import (
	"moonick/internal/model"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppController struct {
	appService *service.AppService
}

func NewAppController(appService *service.AppService) *AppController {
	return &AppController{
		appService: appService,
	}
}

func (app *AppController) GetWebAuth(c *gin.Context) {
	var req model.GetHelloInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(c, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	result, err := app.appService.GetHelloInfo(&req)
	if err != nil {
		zap.L().Error("get hello info failed", zap.Error(err))
		ResponseFailedWithMsg(c, CodeServerBusy, err.Error())
		return
	}

	ResponseSuccess(c, result)
}
