package endpoint

import (
	"idraw-server/api/request"
	"idraw-server/api/response"
	"idraw-server/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func GenerateImageVariantsByImage(c *gin.Context) {

}

func GenerateImagesByImageAndPrompt(c *gin.Context) {

}
