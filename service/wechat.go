package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
)

type weChatLoginResp struct {
	SessionKey string `json:"session_key"`
	OpenId     string `json:"openid"`
	UnionId    string `json:"unionid"`
	ErrCode    string `json:"errcode"`
	ErrMsg     int32  `json:"errmsg"`
}

const weChatApiUrl = "https://api.weixin.qq.com"

func init() {
	validateWechatServiceEnvInjections()
}

func validateWechatServiceEnvInjections() {
	log.Println("validating wechat service's env injections")
	if val := os.Getenv("WE_APP_ID"); val == "" {
		log.Fatalln("lack env WE_APP_ID")
	}
	if val := os.Getenv("WE_APP_SECRET"); val == "" {
		log.Fatalln("lack env WE_APP_SECRET")
	}
	log.Println("validation done")
}

func getWeAppId() string {
	return os.Getenv("WE_APP_ID")
}

func getWeAppSecret() string {
	return os.Getenv("WE_APP_SECRET")
}

func WeChatLogin(code string) (*weChatLoginResp, error) {
	params := url.Values{}
	params.Add("appid", getWeAppId())
	params.Add("secret", getWeAppSecret())
	params.Add("js_code", code)
	params.Add("grant_type", "authorization_code")
	reqUrl := weChatApiUrl + "/sns/jscode2session?" + params.Encode()
	r, err := http.Get(reqUrl)
	result := &weChatLoginResp{}
	if err != nil {
		log.Println("build http request failed", err)
		return result, err
	}
	if r.StatusCode != 200 {
		log.Printf("do http request failed, status: %s", r.Status)
		return result, errors.New(r.Status)
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(result)
	if err != nil {
		log.Println("do response json decode failed", err)
		return result, err
	}
	return result, nil
}
