package http

import (
	"io"
	"encoding/json"
	"net/http"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"errors"
	"mime/multipart"
)

func MakePostReq(url string, postData map[string]interface{}, contentType string) (map[string]interface{}, error) {
	var data io.Reader
	jsonData ,jsonErr := json.Marshal(postData)
	if jsonErr != nil {
		return map[string]interface{}{}, jsonErr
	}
	data = bytes.NewBuffer(jsonData)

	res, _ := http.Post(url, contentType, data)

	var (
		reader io.ReadCloser
		err error
	)
	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(res.Body)
		if err != nil {
			return map[string]interface{}{}, err
		}
	} else {
		reader = res.Body
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return map[string]interface{}{}, err
	}

	var resJsonData map[string]interface{}
	err = json.Unmarshal(body, &resJsonData)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return resJsonData, nil
}


func MakeGetReq(url string, data map[string]string) (map[string]interface{}, error) {

	var count = 0
	for k, v := range data {
		if count == 0 {
			url += "?" + k + "=" + v
		} else {
			url += "&" + k + "=" + v
		}
		count++
	}

	res, err := http.Get(url)
	if err != nil {
		return map[string]interface{}{}, err
	}
	if res.StatusCode != 200 {
		return map[string]interface{}{}, errors.New("网络错误")
	}

	var reader io.ReadCloser
	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(res.Body)
		if err != nil {
			return map[string]interface{}{}, err
		}
	} else {
		reader = res.Body
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return map[string]interface{}{}, err
	}

	var resJsonData map[string]interface{}
	err = json.Unmarshal(body, &resJsonData)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return resJsonData, nil
}

func MakePostFormReq(url string, data map[string]string) (map[string]interface{}, error) {
	buf := new(bytes.Buffer)

	writer := multipart.NewWriter(buf)
	for k,v := range data {
		writer.WriteField(k, v)
	}

	//part, err := writer.CreateFormFile("tmp.png", "tmp.png")
	//if err != nil {
	//	return map[string]interface{}{}, err
	//}
	//
	//fileData := []byte("hello,world")  // 此处内容可以来自本地文件读取或云存储
	//part.Write(fileData)
	//
	//if err = writer.Close(); err != nil {
	//	return map[string]interface{}{}, err
	//}

	req, err := http.NewRequest(http.MethodPost,
		url,
		buf)
	if err != nil {
		return map[string]interface{}{}, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]interface{}{}, err
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if err != nil {
		return map[string]interface{}{}, err
	}

	var resJsonData map[string]interface{}
	err = json.Unmarshal(body, &resJsonData)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return resJsonData, nil
}