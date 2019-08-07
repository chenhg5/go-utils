package upyun

import (
	"github.com/upyun/go-sdk/upyun"
)

var up *upyun.UpYun

func InitUpyun(bucket, operator, pwd string) {
	up = upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   bucket,
		Operator: operator,
		Password: pwd,
	})
}

func Upload(file string, local string) error {

	// 上传文件
	return up.Put(&upyun.PutObjectConfig{
		Path:      file,
		LocalPath: local,
	})
}

func PreProcess(file, local, specification string) error {
	apps := []map[string]interface{}{
		{
			"name":           "thumb",
			"x-gmkerl-thumb": "/crop/" + specification,
			"save_as":        file,
		},
	}

	return asyncPreProcess(local, file, apps)
}

func asyncPreProcess(localPath string, saveKey string, apps []map[string]interface{}) error {

	_, err := up.FormUpload(&upyun.FormUploadConfig{
		LocalPath:      localPath,
		SaveKey:        saveKey,
		NotifyUrl:      "",
		ExpireAfterSec: 60,
		Apps:           apps,
	})
	return err
}

const (
	SpecificationOne   = "700x700a27a176"
	SpecificationTwo   = "1514x1514a44a372"
	SpecificationThree = "1020x1020a0a230"
)
