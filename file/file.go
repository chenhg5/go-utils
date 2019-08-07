package file

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func SaveFile(url string) string {
	path := strings.Split(url, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	}
	out, _ := os.Create("./tmp/" + name)
	defer func() {
		if err := out.Close(); err != nil {
			fmt.Println("SaveFile err", err)
		}
	}()
	resp, _ := http.Get(url)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("SaveFile err2", err)
		}
	}()
	pix, _ := ioutil.ReadAll(resp.Body)
	_, _ = io.Copy(out, bytes.NewReader(pix))
	return name
}
