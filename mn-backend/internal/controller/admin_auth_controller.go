package controller

import (
	"errors"

	"moonick/internal/model/request"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminAuthController struct {
	authService *service.AuthService
	jwtManager  *jwtpkg.Manager
}

func NewAdminAuthController(authService *service.AuthService, jwtManager *jwtpkg.Manager) *AdminAuthController {
	return &AdminAuthController{
		authService: authService,
		jwtManager:  jwtManager,
	}
}

func (c *AdminAuthController) Login(ctx *gin.Context) {
	var req request.AdminLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.authService.AdminLogin(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAdminCredentials):
			ResponseFailedWithMsg(ctx, CodeInvalidPassword, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, resp)
}

func (c *AdminAuthController) Refresh(ctx *gin.Context) {
	token, err := jwtpkg.ExtractBearerToken(ctx.GetHeader("Authorization"))
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, CodeNeedLogin.Msg())
		return
	}

	resp, err := c.authService.RefreshAdminToken(ctx, token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			ResponseFailedWithMsg(ctx, CodeInvalidToken, CodeInvalidToken.Msg())
		case errors.Is(err, service.ErrAdminNotFound):
			ResponseFailedWithMsg(ctx, CodeUserNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, resp)
}

func (c *AdminAuthController) Me(ctx *gin.Context) {
	adminID, err := currentAdminID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}

	admin, err := c.authService.AdminProfile(ctx, adminID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminNotFound):
			ResponseFailedWithMsg(ctx, CodeUserNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, admin)
}
