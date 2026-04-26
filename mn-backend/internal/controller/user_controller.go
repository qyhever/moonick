package controller

import (
	"errors"

	"moonick/internal/model/request"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
	jwtManager  *jwtpkg.Manager
}

func NewUserController(userService *service.UserService, jwtManager *jwtpkg.Manager) *UserController {
	return &UserController{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

func (c *UserController) GetProfile(ctx *gin.Context) {
	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}

	profile, err := c.userService.GetProfile(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			ResponseFailedWithMsg(ctx, CodeUserNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, profile)
}

func (c *UserController) UpdateProfile(ctx *gin.Context) {
	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}

	var req request.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	if err := c.userService.UpdateProfile(ctx, userID, req); err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyNickname):
			ResponseFailedWithMsg(ctx, CodeInvalidParam, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			ResponseFailedWithMsg(ctx, CodeUserNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, gin.H{"ok": true})
}

func (c *UserController) UpdateContact(ctx *gin.Context) {
	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}

	var req request.UpdateContactRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	if err := c.userService.UpdateContact(ctx, userID, req); err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyContact):
			ResponseFailedWithMsg(ctx, CodeInvalidParam, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			ResponseFailedWithMsg(ctx, CodeUserNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, gin.H{"ok": true})
}
