package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"idraw-server/api/request"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/robfig/cron/v3"
)

type GenerationResp struct {
	Created int64            `json:"created"`
	Data    []GenerationData `json:"data"`
}

type GenerationData struct {
	Url string `json:"url"`
}

const openAiApiUrl string = "https://chat-gpt-proxy.danchaofan.xyz/v1/images"

var userUsagesMap = make(map[string]int)

func init() {
	log.Println("fire a cron worker to clean the map")
	c := cron.New()
	c.AddFunc("@daily", func() {
		log.Println("start to clean the map")
		for k := range userUsagesMap {
			delete(userUsagesMap, k)
		}
		log.Println("finished cleaning the map")
	})
	c.Start()
}

// 从 env 中获取，key 属于敏感信息，将会在运行中注入
func getOpenAiApiKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

func GetDailyLimits() int {
	stringValue := os.Getenv("DAILY_LIMITS")
	dailyLimits, _ := strconv.Atoi(stringValue)
	return dailyLimits
}

func GetCurrentUsages(user string) int {
	usages := userUsagesMap[user]
	log.Printf("current user %s, current usages: %d\n", user, usages)
	return usages
}

// GenerateImagesByPrompt 根据场景描述产出符合场景的图片
func GenerateImagesByPrompt(req request.ImageGenerationReq) ([]string, error) {
	usages := GetCurrentUsages(req.User)
	if usages >= GetDailyLimits() {
		return nil, errors.New("current user has exceeded daily limits")
	}
	body, _ := json.Marshal(req)
	r, err := http.NewRequest("POST", openAiApiUrl+"/generations", bytes.NewBuffer(body))
	if err != nil {
		log.Println("build http request failed", err)
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", "Bearer "+getOpenAiApiKey())
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Println("do http request failed, err: ", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("do http request failed, status: %s, body: %s\n", resp.Status, string(b))
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	result := &GenerationResp{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		log.Println("do response json decode failed", err)
		return nil, err
	}
	urls := make([]string, len(result.Data))
	for i, data := range result.Data {
		urls[i] = data.Url
	}
	userUsagesMap[req.User] = usages + 1
	return urls, nil
}

// GenerateImageVariationsByImage 根据图片产出相应变体图片
func GenerateImageVariationsByImage(req request.ImageVariationReq) ([]string, error) {
	usages := GetCurrentUsages(req.User)
	if usages >= GetDailyLimits() {
		return nil, errors.New("current user has exceeded daily limits")
	}
	image, _ := req.File.Open()
	buf := new(bytes.Buffer)
	mp := multipart.NewWriter(buf)
	filePart, _ := mp.CreateFormFile("image", req.File.Filename)
	io.Copy(filePart, image)
	mp.WriteField("user", req.User)
	mp.WriteField("size", req.Size)
	mp.WriteField("n", strconv.Itoa(req.N))
	mp.Close()
	r, err := http.NewRequest("POST", openAiApiUrl+"/variations", buf)
	if err != nil {
		log.Println("build http request failed", err)
		return nil, err
	}
	r.Header.Add("Content-Type", mp.FormDataContentType())
	r.Header.Add("Authorization", "Bearer"+getOpenAiApiKey())
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Println("do http request failed, err: ", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("do http request failed, status: %s, body: %s\n", resp.Status, string(b))
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	result := &GenerationResp{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		log.Println("do response json decode failed", err)
		return nil, err
	}
	urls := make([]string, len(result.Data))
	for i, data := range result.Data {
		urls[i] = data.Url
	}
	userUsagesMap[req.User] = usages + 1
	return urls, nil
}
