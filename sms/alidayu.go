package sms

import (
	"fmt"
	dayu "github.com/gwpp/alidayu-go"
	"github.com/gwpp/alidayu-go/request"
)

type AlidayuSmSType struct {
	AppKey          string
	AppSecret       string
	SmsFreeSignName string // 短信签名
	SmsTemplateCode string
}

func InitAlidayu(key string, secret string, sign string, code string) *AlidayuSmSType {
	var alidayu AlidayuSmSType
	alidayu.AppKey = key
	alidayu.AppSecret = secret
	alidayu.SmsFreeSignName = sign
	alidayu.SmsTemplateCode = code
	return &alidayu
}

func (dayusms *AlidayuSmSType) SendAlidayuSMS(phone string, host string) bool {
	client := dayu.NewTopClient(dayusms.AppKey, dayusms.AppSecret)
	req := request.NewAlibabaAliqinFcSmsNumSendRequest()
	req.SmsFreeSignName = dayusms.SmsFreeSignName
	req.RecNum = phone
	req.SmsTemplateCode = dayusms.SmsTemplateCode
	req.SmsParam = `{"name":"` + host + `"}`
	res, err := client.Execute(req)
	fmt.Println(res)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}