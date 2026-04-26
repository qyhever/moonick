package controller

import (
	"errors"

	"moonick/internal/model/request"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminUserController struct {
	adminService *service.AdminService
}

func NewAdminUserController(adminService *service.AdminService) *AdminUserController {
	return &AdminUserController{adminService: adminService}
}

func (c *AdminUserController) List(ctx *gin.Context) {
	var req request.ListUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.adminService.ListUsers(ctx, req)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *AdminUserController) Detail(ctx *gin.Context) {
	userID, ok := parseInt64Param(ctx, "id")
	if !ok {
		return
	}

	resp, err := c.adminService.GetUserDetail(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			ResponseFailedWithMsg(ctx, CodeUserNotExist, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *AdminUserController) Trips(ctx *gin.Context) {
	userID, ok := parseInt64Param(ctx, "id")
	if !ok {
		return
	}

	var req request.ListTripRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.adminService.ListUserTrips(ctx, userID, req)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(ctx, resp)
}
