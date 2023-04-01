package service

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

const WeChatApiUrl = "https://api.weixin.qq.com"

type WeChatLoginResp struct {
	SessionKey string `json:"session_key"`
	OpenId     string `json:"openid"`
	UnionId    string `json:"unionid"`
	ErrCode    string `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func getWeAppId() string {
	return os.Getenv("WE_APP_ID")
}

func getWeAppSecret() string {
	return os.Getenv("WE_APP_SECRET")
}

func WeChatLogin(code string) (*WeChatLoginResp, error) {
	params := url.Values{}
	params.Add("appid", getWeAppId())
	params.Add("secret", getWeAppSecret())
	params.Add("js_code", code)
	params.Add("grant_type", "authorization_code")
	reqUrl := WeChatApiUrl + "/sns/jscode2session" + params.Encode()
	r, err := http.Get(reqUrl)
	result := &WeChatLoginResp{}
	if err != nil {
		log.Println("build http request failed", err)
		return result, err
	}
	err = json.NewDecoder(r.Body).Decode(result)
	if err != nil {
		log.Println("do response json decode failed", err)
		return result, err
	}
	return result, nil
}
