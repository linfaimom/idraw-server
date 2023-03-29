package request

type ImageGenerationReq struct {
	User   string `json:"user" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
	N      int8   `json:"n" binding:"gte=1,lte=10"`
	Size   string `json:"size" binding:"required"`
}
