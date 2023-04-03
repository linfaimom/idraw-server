package endpoint

import (
	"errors"
	"idraw-server/api/request"
	"idraw-server/api/response"
	"idraw-server/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDailyLimits(c *gin.Context) {
	response.Success(c, service.GetDailyLimits())
}

func GetCurrentUsages(c *gin.Context) {
	if openId := c.Query("openId"); openId != "" {
		data := service.GetCurrentUsages(openId)
		response.Success(c, data)
	} else {
		response.Fail(c, http.StatusBadRequest, errors.New("failed to fetch current usages"))
		return
	}
}

func GenerateImagesByPrompt(c *gin.Context) {
	req := request.ImageGenerationReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err)
		return
	}
	result, err := service.GenerateImagesByPrompt(req)
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, err)
		return
	}
	response.Success(c, result)
}

func GenerateImageVariationsByImage(c *gin.Context) {
	req := request.ImageVariationReq{}
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err)
		return
	}
	result, err := service.GenerateImageVariationsByImage(req)
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, err)
		return
	}
	response.Success(c, result)
}
