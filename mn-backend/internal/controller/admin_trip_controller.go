package controller

import (
	"errors"

	"moonick/internal/model/request"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminTripController struct {
	adminService *service.AdminService
}

func NewAdminTripController(adminService *service.AdminService) *AdminTripController {
	return &AdminTripController{adminService: adminService}
}

func (c *AdminTripController) DashboardSummary(ctx *gin.Context) {
	resp, err := c.adminService.GetDashboardSummary(ctx)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *AdminTripController) List(ctx *gin.Context) {
	var req request.ListTripRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.adminService.ListTrips(ctx, req)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *AdminTripController) Detail(ctx *gin.Context) {
	tripID, ok := parseInt64Param(ctx, "id")
	if !ok {
		return
	}

	resp, err := c.adminService.GetTripDetail(ctx, tripID)
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

func (c *AdminTripController) Update(ctx *gin.Context) {
	tripID, ok := parseInt64Param(ctx, "id")
	if !ok {
		return
	}

	var req request.AdminUpdateTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.adminService.UpdateTrip(ctx, tripID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTripNotFound):
			ResponseFailedWithMsg(ctx, CodeResourceNotExist, err.Error())
		case errors.Is(err, service.ErrTripStatusInvalid):
			ResponseFailedWithMsg(ctx, CodeInvalidParam, err.Error())
		default:
			ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		}
		return
	}
	ResponseSuccess(ctx, resp)
}
