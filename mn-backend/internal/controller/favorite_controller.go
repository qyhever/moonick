package controller

import (
	"errors"

	"moonick/internal/model/request"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type FavoriteController struct {
	favoriteService *service.FavoriteService
	jwtManager      *jwtpkg.Manager
}

func NewFavoriteController(favoriteService *service.FavoriteService, jwtManager *jwtpkg.Manager) *FavoriteController {
	return &FavoriteController{
		favoriteService: favoriteService,
		jwtManager:      jwtManager,
	}
}

func (c *FavoriteController) Toggle(ctx *gin.Context) {
	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}
	paramKey := "tripId"
	if ctx.Param(paramKey) == "" {
		paramKey = "id"
	}
	tripID, ok := parseInt64Param(ctx, paramKey)
	if !ok {
		return
	}

	resp, err := c.favoriteService.Toggle(ctx, userID, tripID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTripNotFound):
			ResponseFailedWithMsg(ctx, CodeResourceNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *FavoriteController) List(ctx *gin.Context) {
	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}

	var req request.ListTripRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.favoriteService.ListFavorites(ctx, userID, req)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(ctx, resp)
}
