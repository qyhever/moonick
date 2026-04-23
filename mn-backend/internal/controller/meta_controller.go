package controller

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type MetaController struct{}

func NewMetaController() *MetaController {
	return &MetaController{}
}

func (mc *MetaController) GetMeta(c *gin.Context) {
	metaData, err := os.ReadFile("./public/meta.json")
	if err != nil {
		// 如果文件不存在，返回默认时间或错误
		c.JSON(http.StatusOK, gin.H{
			"deployTime": "unknown",
		})
		return
	}

	var metaMap map[string]interface{}
	if err := json.Unmarshal(metaData, &metaMap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse meta.json",
		})
		return
	}

	c.JSON(http.StatusOK, metaMap)
}
