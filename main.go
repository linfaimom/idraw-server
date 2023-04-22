package main

import (
	"fmt"
	"idraw-server/api/endpoint"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

const addr = ":8388"

func customLogFormatter(param gin.LogFormatterParams) string {
	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
		param.ClientIP,
		param.TimeStamp.Format(time.RFC1123),
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency,
		param.Request.UserAgent(),
		param.ErrorMessage,
	)
}

func main() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: customLogFormatter,
		SkipPaths: []string{"/ping"},
	}))
	// health check endpoint
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.String(200, "pong")
	})
	// wechat endpoints
	wx := r.Group("/api/wx")
	{
		wx.GET("/login", endpoint.WeLogin)
	}
	// biz service endpoints
	app := r.Group("/api/images")
	{
		// 获取每日限额
		app.GET("/dailyLimits", endpoint.GetDailyLimits)
		// 增加每日限额
		app.POST("/dailyLimits", endpoint.IncreaseDailyLimits)
		// 当前使用值
		app.GET("/currentUsages", endpoint.GetCurrentUsages)
		// 文件下载
		app.GET("", endpoint.ServeFile)
		// 生成记录查询
		app.GET("/records", endpoint.FetchRecords)
		// 生成纪录总数查询
		app.GET("/records/count", endpoint.FetchRecordsCount)
		// 文件上传
		app.POST("", endpoint.UploadFile)
		// 根据场景描述产出符合场景的图片
		app.POST("/generations", endpoint.GenerateImagesByPrompt)
		// 根据图片产出其变体
		app.POST("/variations", endpoint.GenerateImageVariationsByImage)
	}
	if err := r.Run(addr); err != nil {
		log.Println("server start up failed")
	}
}
