package sms

import (
	"fmt"
	dayu "github.com/gwpp/alidayu-go"
	"github.com/gwpp/alidayu-go/request"
)

type AlidayuSmSType struct {
	APP_KEY            string
	APP_SECRET         string
	SMS_FREE_SIGN_NAME string // 短信签名
	SMS_TEMPLATE_CODE  string
}

var Alidayu AlidayuSmSType

func InitAlidayu(key string, secret string, sign string, code string)  {
	Alidayu.APP_KEY = key
	Alidayu.APP_SECRET = secret
	Alidayu.SMS_FREE_SIGN_NAME = sign
	Alidayu.SMS_TEMPLATE_CODE = code
}

func (dayusms AlidayuSmSType) SendAlidayuSMS(phone string, host string) bool {
	client := dayu.NewTopClient(dayusms.APP_KEY, dayusms.APP_SECRET)
	req := request.NewAlibabaAliqinFcSmsNumSendRequest()
	req.SmsFreeSignName = dayusms.SMS_FREE_SIGN_NAME
	req.RecNum = phone
	req.SmsTemplateCode = dayusms.SMS_TEMPLATE_CODE
	req.SmsParam = `{"name":"` + host + `"}`
	res, err := client.Execute(req)
	fmt.Println(res)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}