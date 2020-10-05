package request

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Client define http.Client
func Client(p string) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:    30,
		IdleConnTimeout: time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // lgtm [go/disabled-certificate-check]
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 30,
			KeepAlive: time.Second,
		}).DialContext,
	}

	if p != "" {
		if p, err := url.Parse(p); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	re := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &http.Client{
		Transport:     tr,
		CheckRedirect: re,
		Timeout:       time.Second * 30,
	}
}
