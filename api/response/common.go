package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type RespBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, RespBody{
		Code: http.StatusOK,
		Msg:  "Succeed",
		Data: data,
	})
}

func Fail(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, RespBody{
		Code: statusCode,
		Msg:  err.Error(),
	})
}
