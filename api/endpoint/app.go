package endpoint

import (
	"errors"
	"idraw-server/api/request"
	"idraw-server/api/response"
	"idraw-server/service"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDailyLimits(c *gin.Context) {
	openId := c.Query("openId")
	response.Success(c, service.GetDailyLimits(openId))
}

func GetCurrentUsages(c *gin.Context) {
	if openId := c.Query("openId"); openId != "" {
		data := service.GetCurrentUsages(openId)
		response.Success(c, data)
	} else {
		response.Fail(c, http.StatusBadRequest, errors.New("error params"))
		return
	}
}

func FetchRecordsCount(c *gin.Context) {
	openId := c.Query("openId")
	if openId == "" {
		response.Fail(c, http.StatusBadRequest, errors.New("error params"))
		return
	}
	count, err := service.FetchRecordsCount(openId)
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, err)
		return
	}
	response.Success(c, count)
}

func FetchRecords(c *gin.Context) {
	openId := c.Query("openId")
	calledType := c.Query("calledType")
	if openId == "" || calledType == "" {
		response.Fail(c, http.StatusBadRequest, errors.New("error params"))
		return
	}
	records, err := service.FetchRecords(openId, calledType)
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, err)
		return
	}
	response.Success(c, records)
}

func ServeFile(c *gin.Context) {
	if fileName := c.Query("fileName"); fileName != "" {
		file, err := service.ServeFile(fileName)
		if err != nil {
			response.Fail(c, http.StatusServiceUnavailable, err)
			return
		}
		defer file.Close()
		c.Writer.Header().Add("Content-Type", "image/png")
		_, err = io.Copy(c.Writer, file)
		if err != nil {
			response.Fail(c, http.StatusServiceUnavailable, err)
			return
		}
		response.Success(c, nil)
	} else {
		response.Fail(c, http.StatusBadRequest, errors.New("error params"))
		return
	}
}

func UploadFile(c *gin.Context) {
	req := request.FileUploadReq{}
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err)
		return
	}
	result, err := service.UploadFile(req)
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, err)
		return
	}
	response.Success(c, result)
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
