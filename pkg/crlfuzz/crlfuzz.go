package crlfuzz

import (
	"errors"
	"net/http"
	"strings"

	"dw1.io/crlfuzz/pkg/request"
)

// Scan will scanning for CRLF vulnerability against target
func Scan(url string, method string, data string, headers []string, proxy string) (bool, error) {
	client := request.Client(proxy)
	req, e := http.NewRequest(method, url, strings.NewReader(data))
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)

		if len(parts) != 2 {
			continue
		}

		req.Header.Set(parts[0], parts[1])
	}
	if e != nil {
		return false, errors.New(e.Error())
	}

	res, e := client.Do(req)
	if e != nil {
		return false, errors.New(e.Error())
	}
	defer res.Body.Close()

	return isVuln(res), nil
}

func isVuln(r *http.Response) bool {
	for key, header := range r.Header {
		if key == "Set-Cookie" {
			for _, value := range header {
				if strings.Contains(value, cookie) {
					return true
				}
			}
		}
	}
	return false
}
