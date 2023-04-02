package endpoint

import (
	"errors"
	"idraw-server/api/response"
	"idraw-server/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WeLogin(c *gin.Context) {
	if code := c.Query("code"); code != "" {
		data, err := service.WeChatLogin(code)
		if err != nil {
			response.Fail(c, http.StatusServiceUnavailable, err)
		}
		response.Success(c, *data)
	} else {
		response.Fail(c, http.StatusBadRequest, errors.New("failed to fetch wx code"))
		return
	}
}
