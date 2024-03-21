# Usage Example Http Request
- Example GET
```
func (u *useCase) CheckIP(ctx context.Context) (ipAddress string, asNumber int64) {
	clientRealIP := ctx.Value("ClientRealIP")
	var body map[string]interface{}
	result := httprq.Get(config.AppConfig.CheckRealIP+
		clientRealIP.(string)).WithContext(ctx).
		WithTimeoutHystrix(
			config.AppConfig.HystrixTimeout,
			config.AppConfig.HistrixMaxConcurrentRequests,
			config.AppConfig.HystrixErrorPercentThreshold).
		Execute().Consume(&body)

	bodyAsNumber, ok := body["as_number"].(float64)
	if !ok {
		fmt.Println("err get AS Number")
	}
	asNumber = int64(bodyAsNumber)

	ipAddress, ok = body["ip_address"].(string)
	if !ok {
		fmt.Println("err get IP")
	}

	return ipAddress, asNumber
}
```

```
func (u *useCase) ValidateRechaptcha(ctx context.Context, captcha, platform string) error {
	var secret string
	type SiteVerifyResponse struct {
		Success     bool      `json:"success"`
		Score       float64   `json:"score"`
		Action      string    `json:"action"`
		ChallengeTS time.Time `json:"challenge_ts"`
		Hostname    string    `json:"hostname"`
		ErrorCodes  []string  `json:"error-codes"`
	}

	switch platform {
	case "android":
		secret = config.AppConfig.RecaptchaSecretKeyAndroid
	case "ios":
		secret = config.AppConfig.RecaptchaSecretKeyIos
	case "web":
		secret = config.AppConfig.RecaptchaSecretKeyWeb
	}

	params := make(map[string]string)
	params["secret"] = secret
	params["response"] = captcha

	var body SiteVerifyResponse
	result := httprq.Get(config.AppConfig.RecaptchaSiteVerify).
		AddQueryParams(params).
		WithContext(ctx).
		WithTimeoutHystrix(
			config.AppConfig.HystrixTimeout,
			config.AppConfig.HistrixMaxConcurrentRequests,
			config.AppConfig.HystrixErrorPercentThreshold).
		Execute().Consume(&body)

	// Check recaptcha verification success.
	if !body.Success {
		return errors.New("unsuccessful recaptcha verify request")
	}

	// Check response score.
	if body.Score < 0.8 {
		return errors.New("lower received score than expected")
	}

	return nil
}
```
- Example POST
```
var res MoengagePushAPIResponse
	result := api.Post(url).
		AddHeader("Content-Type", "application/json").
		WithRetryStrategy(api.NewRetryAllErrors()).
		WithBody(bytes.NewBuffer(body)).
		WithTimeout(int(timeout)).
		WithContext(ctx).
		Execute()
 
	if err := result.Consume(&res); err != nil {
		return nil, im.handleError(err)
	}
```

```
func (im *integrationModule) PushEvent(ctx context.Context, req MoengageEventRequest) error {
	const (
		processName = "integration.moengage.PushEvent"
	)
	span, _ := opentracing.StartSpanFromContext(ctx, processName)
	defer span.Finish()
 
	moengageCfg := im.endpoints.Moengage
	body, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/%s/%s",
		moengageCfg.BaseUrl,
		moengageCfg.BasePathEventAPI,
		moengageCfg.AppID)
	timeout := time.Duration(moengageCfg.Timeout).Seconds()
 
	var res MoengageBodyResponse
	err := api.Post(url).
		SetBasicAuth(moengageCfg.AppID, moengageCfg.AppKey).
		AddHeader("Content-Type", "application/json").
		AddHeader("MOE-APPKEY", moengageCfg.MoeAppKey).
		WithRetryStrategy(api.NewRetryAllErrors()).
		WithBody(bytes.NewBuffer(body)).
		WithTimeout(int(timeout)).
		WithContext(ctx).
		Execute().Consume(&res)
 
	log.WithFields(log.Fields{
		"appID":     moengageCfg.AppID,
		"appKey":    moengageCfg.AppKey,
		"appSecret": moengageCfg.MoeAppKey,
		"request":   req,
		"response":  res,
	}).Infoln("Logging moengage...")
 
	return err
}
```
```
func (im *integrationModule) PushNotification(ctx context.Context, req MoengagePushAPIRequest) (*MoengagePushAPIResponse, error) {
	const (
		processName = "integration.moengage.PushNotification"
	)
	span, _ := opentracing.StartSpanFromContext(ctx, processName)
	defer span.Finish()
 
	moengageCfg := im.endpoints.Moengage
 
	req.AppID = moengageCfg.AppID
	req.Signature = im.generateSignature(
		ctx,
		moengageCfg.AppID,
		req.CampaignName,
		moengageCfg.MoeAppKey,
	)
 
	body, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/%s",
		moengageCfg.BaseUrl,
		moengageCfg.BasePathPushAPI)
 
	timeout := time.Duration(moengageCfg.Timeout).Seconds()
 
	var res MoengagePushAPIResponse
	result := api.Post(url).
		AddHeader("Content-Type", "application/json").
		WithRetryStrategy(api.NewRetryAllErrors()).
		WithBody(bytes.NewBuffer(body)).
		WithTimeout(int(timeout)).
		WithContext(ctx).
		Execute()
 
	if err := result.Consume(&res); err != nil {
		return nil, im.handleError(err)
	}
 
	return &res, nil
}
```

## Feature
- [x] Hystrix Timeout
- [x] Http Timeout
- [x] Retry Mechanism
- [ ] add singleflag