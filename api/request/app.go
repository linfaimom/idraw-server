package request

import "mime/multipart"

type FileUploadReq struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
	User string                `form:"user" binding:"required"`
}

type ImageGenerationReq struct {
	Model  string `json:"model" binding:"required"`
	User   string `json:"user" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
	N      int    `json:"n" binding:"gte=1,lte=10"`
	Size   string `json:"size" binding:"required"`
}

type ImageVariationReq struct {
	FilePath string `form:"filePath" binding:"required"`
	User     string `form:"user" binding:"required"`
	N        int    `form:"n" binding:"gte=1,lte=10"`
	Size     string `form:"size" binding:"required"`
}
