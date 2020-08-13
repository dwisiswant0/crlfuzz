package crlfuzz

import "strings"

// GenerateURL should generate for potential vulnerability URLs
func GenerateURL(u string) []string {
	var url []string

	if !strings.HasSuffix(u, "/") {
		u += "/"
	}

	for _, a := range appendList {
		for _, e := range escapeList {
			url = append(url, u+a+e+injector)
		}
	}

	return url
}
