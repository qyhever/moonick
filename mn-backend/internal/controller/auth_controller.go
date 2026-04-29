package controller

import (
	"errors"
	"strconv"

	"moonick/internal/model/request"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
	userService *service.UserService
	jwtManager  *jwtpkg.Manager
}

func NewAuthController(authService *service.AuthService, userService *service.UserService, jwtManager *jwtpkg.Manager) *AuthController {
	return &AuthController{
		authService: authService,
		userService: userService,
		jwtManager:  jwtManager,
	}
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req request.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.authService.Register(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPhoneAlreadyRegistered):
			ResponseFailedWithMsg(ctx, CodeUserExist, err.Error())
		case errors.Is(err, service.ErrInvalidUserCredentials):
			ResponseFailedWithMsg(ctx, CodeInvalidParam, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, resp)
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.authService.Login(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserCredentials):
			ResponseFailedWithMsg(ctx, CodeInvalidPassword, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, resp)
}

func (c *AuthController) Refresh(ctx *gin.Context) {
	token, err := jwtpkg.ExtractBearerToken(ctx.GetHeader("Authorization"))
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, CodeNeedLogin.Msg())
		return
	}

	resp, err := c.authService.RefreshUserToken(ctx, token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			ResponseFailedWithMsg(ctx, CodeInvalidToken, CodeInvalidToken.Msg())
		case errors.Is(err, service.ErrUserNotFound):
			ResponseFailedWithMsg(ctx, CodeUserNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, resp)
}

func (c *AuthController) Me(ctx *gin.Context) {
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

func currentUserID(ctx *gin.Context, manager *jwtpkg.Manager) (int64, error) {
	return currentSubjectID(ctx, manager)
}

func currentAdminID(ctx *gin.Context, manager *jwtpkg.Manager) (int64, error) {
	return currentSubjectID(ctx, manager)
}

func currentSubjectID(ctx *gin.Context, manager *jwtpkg.Manager) (int64, error) {
	if claimsValue, ok := ctx.Get(jwtpkg.ContextClaimsKey); ok {
		if claims, ok := claimsValue.(*jwtpkg.Claims); ok && claims != nil {
			return strconv.ParseInt(claims.Subject, 10, 64)
		}
	}

	if manager == nil {
		return 0, errors.New("jwt manager is nil")
	}

	token, err := jwtpkg.ExtractBearerToken(ctx.GetHeader("Authorization"))
	if err != nil {
		return 0, err
	}
	claims, err := manager.Parse(token)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(claims.Subject, 10, 64)
}
