package controller

import (
	"errors"

	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	userService *service.UserService
	jwtManager  *jwtpkg.Manager
}

func NewFileController(userService *service.UserService, jwtManager *jwtpkg.Manager) *FileController {
	return &FileController{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

func (c *FileController) UploadAvatar(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请选择头像文件")
		return
	}

	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}

	if err := c.userService.UpdateAvatar(ctx, userID, file); err != nil {
		switch {
		case errors.Is(err, service.ErrAvatarFileRequired):
			ResponseFailedWithMsg(ctx, CodeInvalidParam, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			ResponseFailedWithMsg(ctx, CodeUserNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
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

	ResponseSuccess(ctx, gin.H{"url": profile.AvatarURL})
}
