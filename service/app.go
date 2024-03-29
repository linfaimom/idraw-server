package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"idraw-server/api/request"
	"idraw-server/api/response"
	"idraw-server/db"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sunshineplan/imgconv"

	"github.com/robfig/cron/v3"
)

type generationResp struct {
	Created int64            `json:"created"`
	Data    []generationData `json:"data"`
}

type generationData struct {
	Url string `json:"url"`
}

type errorResp struct {
	Error errorData `json:"error"`
}

type errorData struct {
	Code    any    `json:"code"`
	Message string `json:"message"`
	Param   any    `json:"param"`
	Type    string `json:"type"`
}

const (
	openAiApiUrl       string = "https://openai.freedom-island.xyz/v1/images"
	dataDir            string = "/data" // mount this dir to the nas for persistence
	uploadedPath       string = "/idraw-uploaded-dir/"
	generatedPath      string = "/idraw-generated-dir/"
	typePrompt         string = "PROMPT"
	typeVariation      string = "VARIATION"
	prefixDailyLimits  string = "limits-"
	prefixCurrentUsage string = "usage-"
)

var (
	ctx          context.Context
	redisCli     *redis.Client
	recordMapper db.RecordMapper
)

func init() {
	validateAppServiceEnvInjections()
	recordMapper = db.NewRecordMapper()
	log.Println("create a redis client")
	ctx = context.Background()
	redisCli = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})
	log.Println("fire a cron worker to reset the redis value")
	c := cron.New()
	c.AddFunc("@daily", func() {
		log.Println("start to reset the redis value")
		// reset the usages
		usageIter := redisCli.Scan(ctx, 0, prefixCurrentUsage+"*", 0).Iterator()
		for usageIter.Next(ctx) {
			redisCli.Set(ctx, usageIter.Val(), 0, 0)
		}
		// reset the limits
		limitsIter := redisCli.Scan(ctx, 0, prefixDailyLimits+"*", 0).Iterator()
		for limitsIter.Next(ctx) {
			redisCli.Set(ctx, limitsIter.Val(), os.Getenv("DAILY_LIMITS"), 0)
		}
		log.Println("finished reseting the redis value")
	})
	c.Start()
}

func validateAppServiceEnvInjections() {
	log.Println("validating app service's env injections")
	if val := os.Getenv("OPENAI_API_KEY"); val == "" {
		log.Fatalln("lack env OPENAI_API_KEY")
	}
	if val := os.Getenv("DAILY_LIMITS"); val == "" {
		log.Fatalln("lack env DAILY_LIMITS")
	}
	if val := os.Getenv("REDIS_ADDR"); val == "" {
		log.Fatalln("lack env REDIS_ADDR")
	}
	log.Println("validation done")
}

// 从 env 中获取，key 属于敏感信息，将会在运行中注入
func getOpenAiApiKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

func GetDailyLimits(user string) int {
	defaultVal := os.Getenv("DAILY_LIMITS")
	if user == "" {
		result, _ := strconv.Atoi(defaultVal)
		return result
	}
	val, err := redisCli.Get(ctx, prefixDailyLimits+user).Result()
	if err != nil {
		if err == redis.Nil {
			redisCli.Set(ctx, prefixDailyLimits+user, defaultVal, 0)
		} else {
			log.Println("failed to get daily limits, set value as default")
		}
		val = defaultVal
	}
	limits, _ := strconv.Atoi(val)
	return limits
}

func IncreaseDailyLimits(user string, num int) {
	val, _ := redisCli.Get(ctx, prefixDailyLimits+user).Result()
	result, _ := strconv.Atoi(val)
	redisCli.Set(ctx, prefixDailyLimits+user, result+num, 0)
}

func GetCurrentUsages(user string) int {
	val, err := redisCli.Get(ctx, prefixCurrentUsage+user).Result()
	if err != nil {
		if err == redis.Nil {
			redisCli.Set(ctx, prefixCurrentUsage+user, 0, 0)
		} else {
			log.Println("failed to get current usage, set value as 0")
		}
		val = "0"
	}
	usages, _ := strconv.Atoi(val)
	return usages
}

// ServeFile 提供文件下载功能
func ServeFile(relativePath string) (*os.File, error) {
	filePath := dataDir + relativePath
	return os.Open(filePath)
}

// UploadFile 接收文件上传，并保存至数据目录（外挂 nas 持久化）
func UploadFile(req request.FileUploadReq) (string, error) {
	file := req.File
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	// for security reasons, we just expose the relative path not the full path to the outside world
	relativeDst := uploadedPath + req.User + "-" + file.Filename
	if err = os.MkdirAll(filepath.Dir(dataDir+relativeDst), 0750); err != nil {
		return "", err
	}
	out, err := os.Create(dataDir + relativeDst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	if err != nil {
		return "", nil
	}
	log.Println("saved file in ", dataDir+relativeDst)
	return relativeDst, nil
}

func FetchRecordsCount(openId string) (int, error) {
	count, err := recordMapper.FetchCountByUser(openId)
	if err != nil && err.Error() != "record not found" {
		log.Printf("fetch user %s's records count failed, the error is %s\n", openId, err)
		return 0, err
	}
	return count, nil
}

func FetchRecords(openId string, calledType string) ([]response.RecordDto, error) {
	if calledType != typePrompt && calledType != typeVariation {
		return []response.RecordDto{}, errors.New("not a valid called type")
	}
	records, err := recordMapper.FetchByUserAndType(openId, calledType)
	if err != nil && err.Error() != "record not found" {
		log.Printf("fetch user %s's records failed, the error is %s\n", openId, err)
		return []response.RecordDto{}, err
	}
	result := make([]response.RecordDto, len(records))
	for i, v := range records {
		var output []string
		_ = json.Unmarshal([]byte(v.Output), &output)
		dto := response.RecordDto{
			Id:     v.ID,
			Type:   v.Type,
			Input:  v.Input,
			Output: output,
		}
		result[i] = dto
	}
	return result, nil
}

// GenerateImagesByPrompt 根据场景描述产出符合场景的图片
func GenerateImagesByPrompt(req request.ImageGenerationReq) ([]string, error) {
	usages := GetCurrentUsages(req.User)
	log.Printf("do prompt request, current user %s, current usages: %d\n", req.User, usages)
	if usages >= GetDailyLimits(req.User) {
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
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		result := &errorResp{}
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			b, _ := io.ReadAll(resp.Body)
			log.Printf("do decode body failed, just print string, status: %s, body: %s\n", resp.Status, string(b))
			return nil, errors.New(resp.Status)
		}
		log.Printf("error response, status is: %s, msg is: %s\n", resp.Status, result.Error.Message)
		return nil, errors.New(result.Error.Message)
	}
	result := &generationResp{}
	json.NewDecoder(resp.Body).Decode(result)
	urls := make([]string, len(result.Data))
	for i, data := range result.Data {
		// download from the url and save as a file
		fileUrl, err := saveFile(typePrompt, req.User, data.Url)
		if err != nil {
			log.Println("save file error")
			return []string{}, err
		}
		urls[i] = fileUrl
	}
	// save record to db
	jsonStr, _ := json.Marshal(urls)
	recordMapper.Insert(req.User, typePrompt, req.Prompt, string(jsonStr))
	// accumulate usages
	accumulateCurrentUsage(req.User)
	return urls, nil
}

// GenerateImageVariationsByImage 根据图片产出相应变体图片
func GenerateImageVariationsByImage(req request.ImageVariationReq) ([]string, error) {
	usages := GetCurrentUsages(req.User)
	log.Printf("do variation request, current user %s, current usages: %d\n", req.User, usages)
	if usages >= GetDailyLimits(req.User) {
		return nil, errors.New("current user has exceeded daily limits")
	}
	fileDst := dataDir + req.FilePath
	file, err := imgconv.Open(fileDst)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	mp := multipart.NewWriter(buf)
	filePart, _ := mp.CreateFormFile("image", "image.png")
	imgconv.Write(filePart, file, &imgconv.FormatOption{Format: imgconv.PNG})
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
	r.Header.Add("Authorization", "Bearer "+getOpenAiApiKey())
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Println("do http request failed, err: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		result := &errorResp{}
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			b, _ := io.ReadAll(resp.Body)
			log.Printf("do decode body failed, just print string, status: %s, body: %s\n", resp.Status, string(b))
			return nil, errors.New(resp.Status)
		}
		log.Printf("error response, status is: %s, msg is: %s\n", resp.Status, result.Error.Message)
		return nil, errors.New(result.Error.Message)
	}
	result := &generationResp{}
	json.NewDecoder(resp.Body).Decode(result)
	urls := make([]string, len(result.Data))
	for i, data := range result.Data {
		// download from the url and save as a file
		fileUrl, err := saveFile(typeVariation, req.User, data.Url)
		if err != nil {
			log.Println("save file error")
			return []string{}, err
		}
		urls[i] = fileUrl
	}
	// save record to db
	jsonStr, _ := json.Marshal(urls)
	recordMapper.Insert(req.User, typeVariation, req.FilePath, string(jsonStr))
	// accumulate usages
	accumulateCurrentUsage(req.User)
	return urls, nil
}

func saveFile(calledType string, user string, url string) (string, error) {
	fileName := fmt.Sprintf("%s-%s-%d.png", user, calledType, time.Now().UnixNano()/int64(time.Millisecond))
	relativeDst := generatedPath + fileName
	if err := os.MkdirAll(filepath.Dir(dataDir+relativeDst), 0750); err != nil {
		return "", err
	}
	out, err := os.Create(dataDir + relativeDst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	// download the content
	resp, err := http.Get(url)
	if err != nil {
		log.Println("do download file request failed, err: ", err)
		return "", err
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("do download file request failed, status: %s, body: %s\n", resp.Status, string(b))
		return "", errors.New(resp.Status)
	}
	defer resp.Body.Close()
	// copy to the file
	size, _ := io.Copy(out, resp.Body)
	log.Printf("save file %s completed, the size is: %d bytes", fileName, size)
	return relativeDst, nil
}

func accumulateCurrentUsage(user string) {
	val, err := redisCli.Get(ctx, prefixCurrentUsage+user).Result()
	if err == redis.Nil {
		redisCli.Set(ctx, prefixCurrentUsage+user, 1, 0)
	} else if err == nil {
		intVar, _ := strconv.Atoi(val)
		redisCli.Set(ctx, prefixCurrentUsage+user, intVar+1, 0)
	} else {
		log.Printf("just ignore this accumulation, the error is %s", err)
	}
}
