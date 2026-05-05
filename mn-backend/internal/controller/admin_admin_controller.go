package controller

import (
	"errors"

	"moonick/internal/model/request"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminAdminController struct {
	adminService *service.AdminService
}

func NewAdminAdminController(adminService *service.AdminService) *AdminAdminController {
	return &AdminAdminController{adminService: adminService}
}

func (c *AdminAdminController) Create(ctx *gin.Context) {
	var req request.CreateAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.adminService.CreateAdmin(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminUsernameRequired),
			errors.Is(err, service.ErrAdminPasswordRequired):
			ResponseFailedWithMsg(ctx, CodeInvalidParam, err.Error())
		case errors.Is(err, service.ErrAdminUsernameAlreadyExists):
			ResponseFailedWithMsg(ctx, CodeResourceExists, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}

	ResponseSuccess(ctx, resp)
}
