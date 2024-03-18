package main

import (
	"bitbucket.org/rctiplus/vegapunk/httprq"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

var (
	BaseUrl = "https://hera.mncplus.id/core-idp/api/v1/token/visitor/"
	apikey  = "43aCSi34YX5wUf4Fd3kb5Lbdvzwyx9f2"
)

type Request struct {
	Platform string `json:"platform"`
	DeviceId string `json:"device_id"`
}

func main() {
	//token := generateToken()
	//fmt.Println(token)
	//validateToken(token)
	LogToFile("main", "main init")
	for i := 0; i < 100; i++ {
		audioChecker()
		//time.Sleep(1)
	}
}

func validateToken(token string) {
	var body map[string]interface{}
	headers := make(map[string]string)
	headers["apikey"] = apikey
	headers["Authorization"] = token

	err := httprq.Get(BaseUrl+"validate").WithContext(context.Background()).
		WithTimeoutHystrix(
			5000,
			100,
			25).
		AddHeaders(headers).
		Execute().Consume(&body)

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	fmt.Println(body)
}

func generateToken() string {
	var body map[string]interface{}
	request := Request{
		Platform: "web",
		DeviceId: "2313232",
	}
	requestBody, _ := json.Marshal(request)

	headers := make(map[string]string)
	headers["apikey"] = apikey
	headers["Content-Type"] = "application/json"

	err := httprq.Post(BaseUrl).WithContext(context.Background()).
		WithTimeoutHystrix(
			5000,
			100,
			25).
		AddHeaders(headers).
		WithBody(bytes.NewBuffer(requestBody)).
		Execute().Consume(&body)

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	fmt.Println(body)
	return body["data"].(map[string]interface{})["access_token"].(string)
}

func audioChecker() {
	var body map[string]interface{}

	url := "https://shazam-api6.p.rapidapi.com/shazam/recognize?url=" + "https://short.rctiplus.id/vod-e5a2a2/3e2e24e0537371eebf8a87c6371c0102/6e875c73eb2049309d14ba001ca61da0-86df56bea1be511422704bafd7057318-sq.mp3"
	//url := "https://rapidapi.akhmadjonov.uz/shazam/recognize/?url=https://short.rctiplus.id/vod-e5a2a2/40f9ef42c3ee71eea26287c7361c0102/2bc3aa0cb7ce454abcfac66b5ba50ecc-9547ae9f13ce620a78371a988ebb25e5-sq.mp3"
	err := httprq.Get(url).
		WithContext(context.Background()).
		AddHeader("X-RapidAPI-Key", "4458aa9519msh0e2d649d2c4a857p1647ecjsnbe52068de9fe").
		AddHeader("X-RapidAPI-Host", "shazam-api6.p.rapidapi.com").
		WithRetryStrategyWhenTimout(3).
		SetErrNotSuccess(false).
		//WithTimeout(3).
		//WithRetryStrategyAllErrors(3).
		//WithTimeoutHystrix(
		//	10000,
		//	10000,
		//	25).
		Execute().
		Consume(&body)

	if err != nil {
		fmt.Println(err.Error())
		LogToFile("error WithRetryStrategyAllErrors", fmt.Sprint(err))
	}

	bodyText, _ := ParseResponseBody(body)
	fmt.Println(bodyText)
}

func ParseResponseBody(body map[string]interface{}) (response string, err error) {
	bodyText, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Error ParseResponseBody :", err)
	}

	return string(bodyText), err
}
