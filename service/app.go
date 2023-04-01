package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"idraw-server/api/request"
	"log"
	"net/http"
	"os"
)

type GenerationResp struct {
	Created int64            `json:"created"`
	Data    []GenerationData `json:"data"`
}

type GenerationData struct {
	Url string `json:"url"`
}

const OpenAiApiUrl = "https://chat-gpt-proxy.danchaofan.xyz/v1/images"

// 从 env 中获取，key 属于敏感信息，将会在运行中注入
func getOpenAiApiKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

// GenerateImagesByPrompt 根据场景描述产出符合场景的图片
func GenerateImagesByPrompt(req request.ImageGenerationReq) ([]string, error) {
	log.Println("the user who submiited this request is ", req.User)
	body, _ := json.Marshal(req)
	r, err := http.Post(OpenAiApiUrl+"/generations", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Println("do http request failed", err)
		return nil, err
	}
	if r.StatusCode != 200 {
		log.Printf("do http request failed, status: %s", r.Status)
		return nil, errors.New(r.Status)
	}
	defer r.Body.Close()
	result := &GenerationResp{}
	err = json.NewDecoder(r.Body).Decode(result)
	if err != nil {
		log.Println("do response json decode failed", err)
		return nil, err
	}
	urls := make([]string, len(result.Data))
	for i, data := range result.Data {
		urls[i] = data.Url
	}
	return urls, nil
}
