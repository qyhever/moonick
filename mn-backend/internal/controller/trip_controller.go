package controller

import (
	"errors"
	"strconv"

	"moonick/internal/model/request"
	jwtpkg "moonick/internal/pkg/jwt"
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
)

type TripController struct {
	tripService *service.TripService
	jwtManager  *jwtpkg.Manager
}

func NewTripController(tripService *service.TripService, jwtManager *jwtpkg.Manager) *TripController {
	return &TripController{
		tripService: tripService,
		jwtManager:  jwtManager,
	}
}

func (c *TripController) List(ctx *gin.Context) {
	var req request.ListTripRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.tripService.ListTrips(ctx, req)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *TripController) Detail(ctx *gin.Context) {
	tripID, ok := parseInt64Param(ctx, "id")
	if !ok {
		return
	}

	resp, err := c.tripService.GetTripDetail(ctx, tripID)
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

func (c *TripController) Create(ctx *gin.Context) {
	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}

	var req request.UpsertTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.tripService.CreateTrip(ctx, userID, req)
	if err != nil {
		handleTripMutationError(ctx, err)
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *TripController) Update(ctx *gin.Context) {
	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}
	tripID, ok := parseInt64Param(ctx, "id")
	if !ok {
		return
	}

	var req request.UpsertTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.tripService.UpdateTrip(ctx, userID, tripID, req)
	if err != nil {
		handleTripMutationError(ctx, err)
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *TripController) ListMine(ctx *gin.Context) {
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

	resp, err := c.tripService.ListMyTrips(ctx, userID, req)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
		return
	}
	ResponseSuccess(ctx, resp)
}

func (c *TripController) UpdateStatus(ctx *gin.Context) {
	userID, err := currentUserID(ctx, c.jwtManager)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeNeedLogin, "登录信息无效")
		return
	}
	tripID, ok := parseInt64Param(ctx, "id")
	if !ok {
		return
	}

	var req request.AdminUpdateTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := c.tripService.UpdateTripStatus(ctx, userID, tripID, req.Status)
	if err != nil {
		handleTripMutationError(ctx, err)
		return
	}
	ResponseSuccess(ctx, resp)
}

func handleTripMutationError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrTripInvalidRoute),
		errors.Is(err, service.ErrTripDepartureInPast),
		errors.Is(err, service.ErrTripContactRequired),
		errors.Is(err, service.ErrTripTypeRequired),
		errors.Is(err, service.ErrTripSeatCountInvalid),
		errors.Is(err, service.ErrTripPriceAmountInvalid),
		errors.Is(err, service.ErrTripStatusInvalid),
		errors.Is(err, service.ErrTripDepartureDateInvalid),
		errors.Is(err, service.ErrTripDepartureTimeInvalid):
		ResponseFailedWithMsg(ctx, CodeInvalidParam, err.Error())
	case errors.Is(err, service.ErrTripNotFound):
		ResponseFailedWithMsg(ctx, CodeResourceNotExist, err.Error())
	case errors.Is(err, service.ErrTripPermissionDenied):
		ResponseFailedWithMsg(ctx, CodePermissionDenied, err.Error())
	default:
		ResponseFailedWithMsg(ctx, CodeServerBusy, err.Error())
	}
}

func parseInt64Param(ctx *gin.Context, key string) (int64, bool) {
	value := ctx.Param(key)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		ResponseFailedWithMsg(ctx, CodeInvalidParam, "请求参数错误: 无效ID")
		return 0, false
	}
	return id, true
}
