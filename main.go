package main

import (
	"idraw-server/api/endpoint"
	"log"

	"github.com/gin-gonic/gin"
)

const addr = ":8388"

func main() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/ping"},
	}))
	// health check endpoint
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.String(200, "pong")
	})
	wx := r.Group("/api/wx")
	{
		wx.POST("/login", func(ctx *gin.Context) {

		})
		wx.POST("/logout", func(ctx *gin.Context) {

		})
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
