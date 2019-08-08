package translate

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
)

func JinShan(word string) string {
	u := "http://fy.iciba.com/ajax.php?a=fy&f=auto&t=auto&w=" + url.QueryEscape(word)
	resp, err := http.Get(u)

	if err != nil {
		panic(err)
	}

	if resp.Body == nil {
		panic("invalid connection")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, _ := ioutil.ReadAll(resp.Body)

	return gjson.Get(string(body), "content.out").String()
}
