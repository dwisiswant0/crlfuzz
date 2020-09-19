package runner

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/dwisiswant0/crlfuzz/pkg/errors"
)

func (o *Options) validate() {
	if isStdin() {
		b, e := ioutil.ReadAll(os.Stdin)
		if e != nil {
			errors.Exit(e.Error())
		}
		o.Target = string(b)
	} else if o.URL != "" {
		o.Target = o.URL
	} else if o.List != "" {
		f, e := ioutil.ReadFile(o.List)
		if e != nil {
			errors.Exit(e.Error())
		}
		o.Target = string(f)
	} else {
		errors.Exit("No target input provided.")
	}

	if o.Saveto != "" {
		f, e := os.OpenFile(o.Saveto,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if e != nil {
			errors.Exit(e.Error())
		}
		o.Output = f
	}
}

func isStdin() bool {
	f, e := os.Stdin.Stat()
	if e != nil {
		return false
	}

	if f.Mode()&os.ModeNamedPipe == 0 {
		return false
	}

	return true
}

func isURL(s string) bool {
	_, e := url.ParseRequestURI(s)
	if e != nil {
		return false
	}

	u, e := url.Parse(s)
	if e != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}
