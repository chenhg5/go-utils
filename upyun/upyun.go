package upyun

import (
	"bytes"
	"fmt"
	"github.com/upyun/go-sdk/upyun"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

func saveFile(url string) string {
	path := strings.Split(url, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	}
	out, _ := os.Create("./tmp/" + name)
	defer func() {
		if err := out.Close(); err != nil {
			fmt.Println("err")
		}
	}()
	resp, _ := http.Get(url)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("err")
		}
	}()
	pix, _ := ioutil.ReadAll(resp.Body)
	_, _ = io.Copy(out, bytes.NewReader(pix))
	return name
}
