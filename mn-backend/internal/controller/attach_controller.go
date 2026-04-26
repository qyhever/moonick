package controller

import (
	"moonick/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AttachController struct {
	service *service.AttachService
}

func NewAttachController() *AttachController {
	s, err := service.NewAttachService()
	if err != nil {
		zap.L().Warn("Failed to initialize AttachService (R2 config might be missing)", zap.Error(err))
		return &AttachController{service: nil}
	}
	return &AttachController{service: s}
}

// Upload 处理文件上传
func (ac *AttachController) Upload(c *gin.Context) {
	if ac.service == nil {
		ResponseFailedWithMsg(c, CodeServerBusy, "R2 service not initialized")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		ResponseFailedWithMsg(c, CodeInvalidParam, "Invalid file: "+err.Error())
		return
	}

	key, err := ac.service.UploadFile(file)
	if err != nil {
		zap.L().Error("Upload failed", zap.Error(err))
		ResponseFailedWithMsg(c, CodeServerBusy, "Upload failed: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Upload successful",
		"key":     key,
	})
}

// Upload 处理文件上传 - test
func (ac *AttachController) Add(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		ResponseFailedWithMsg(c, CodeInvalidParam, "Invalid file: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"message":  "Upload successful",
		"filename": file.Filename,
	})
}

// Delete 处理文件删除
func (ac *AttachController) Delete(c *gin.Context) {
	if ac.service == nil {
		ResponseFailedWithMsg(c, CodeServerBusy, "R2 service not initialized")
		return
	}

	key := c.Query("key")
	if key == "" {
		ResponseFailedWithMsg(c, CodeInvalidParam, "Key is required")
		return
	}

	if err := ac.service.DeleteFile(key); err != nil {
		zap.L().Error("Delete failed", zap.Error(err))
		ResponseFailedWithMsg(c, CodeServerBusy, "Delete failed: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{"message": "Delete successful"})
}

// List 处理文件列表
func (ac *AttachController) List(c *gin.Context) {
	if ac.service == nil {
		ResponseFailedWithMsg(c, CodeServerBusy, "R2 service not initialized")
		return
	}

	bucketName, files, err := ac.service.ListFiles()
	if err != nil {
		zap.L().Error("List failed", zap.Error(err))
		ResponseFailedWithMsg(c, CodeServerBusy, "List failed: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{"files": files, "bucketName": bucketName})
}
