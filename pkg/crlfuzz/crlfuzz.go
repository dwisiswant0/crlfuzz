package crlfuzz

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dwisiswant0/crlfuzz/pkg/request"
)

// Scan will scanning for CRLF vulnerability against target
func Scan(url string, method string, data string, headers []string, proxy string) (bool, error) {
	res, e := request.RawRequest(method, url, data, headers, proxy)
	if e != nil {
		return false, errors.New(e.Error())
	}

	return isVuln(res), nil
}

func isVuln(r *http.Response) bool {
	for key, header := range r.Header {
		if key == keyHeader {
			for _, value := range header {
				if strings.Contains(value, valHeader) {
					return true
				}
			}
		}
	}

	return false
}
