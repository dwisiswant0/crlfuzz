package crlfuzz

import (
	"fmt"
	"strings"
)

// GenerateURL should generate for potential vulnerability URLs
func GenerateURL(u string) []string {
	var url []string

	if !strings.HasSuffix(u, "/") {
		u += "/"
	}

	for _, a := range appendList {
		for _, e := range escapeList {
			url = append(url, fmt.Sprint(u, a, e, keyHeader, "%3a%20", valHeader))
		}
	}

	return url
}
