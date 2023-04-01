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
		// 根据场景描述产出符合场景的图片
		app.POST("/generations", endpoint.GenerateImagesByPrompt)
		// 根据图片产出其变体
		app.POST("/variations", endpoint.GenerateImageVariantsByImage)
		// 根据图片&场景描述产出融入了所需场景的新图片
		app.POST("/edits", endpoint.GenerateImagesByImageAndPrompt)
	}
	if err := r.Run(addr); err != nil {
		log.Println("server start up failed")
	}
}
