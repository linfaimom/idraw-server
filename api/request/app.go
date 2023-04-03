package request

import "mime/multipart"

type ImageGenerationReq struct {
	User   string `json:"user" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
	N      int8   `json:"n" binding:"gte=1,lte=10"`
	Size   string `json:"size" binding:"required"`
}

type ImageVariationReq struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
	User string                `form:"user" binding:"required"`
	N    int8                  `form:"n" binding:"gte=1,lte=10"`
	Size string                `form:"size" binding:"required"`
}
