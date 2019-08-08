package file

import (
	"bytes"
	"fmt"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func SaveFile(url string, isProxy bool) string {
	path := strings.Split(url, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	}
	out, err := os.Create("/tmp/" + name)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := out.Close(); err != nil {
			fmt.Println("SaveFile err", err)
		}
	}()

	var resp *http.Response

	if isProxy {
		dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, proxy.Direct)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "can't connect to the proxy", err)
			return ""
		}
		httpTransport := &http.Transport{}
		httpClient := &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial
		resp, _ = httpClient.Get(url)
	} else {
		resp, _ = http.Get(url)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("SaveFile err2", err)
		}
	}()
	pix, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, bytes.NewReader(pix))

	if err != nil {
		panic(err)
	}

	return name
}
