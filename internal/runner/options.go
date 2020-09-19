package runner

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Headers defines custom headers
type Headers []string

func (h Headers) String() string {
	return strings.Join(h, ", ")
}

// Set defines given each header
func (h *Headers) Set(val string) error {
	*h = append(*h, val)
	return nil
}

// Options defines user args
type Options struct {
	URL, List, Method, Data, Proxy, Target, Saveto string
	Concurrency                                    int
	Version, Silent, Verbose                       bool
	Headers                                        Headers
	Output                                         *os.File
}

var o *Options

func init() {
	o = &Options{}

	flag.StringVar(&o.URL, "u", "", "")
	flag.StringVar(&o.URL, "url", "", "")

	flag.StringVar(&o.List, "l", "", "")
	flag.StringVar(&o.List, "list", "", "")

	flag.StringVar(&o.Method, "X", "GET", "")
	flag.StringVar(&o.Method, "method", "GET", "")

	flag.StringVar(&o.Saveto, "o", "", "")
	flag.StringVar(&o.Saveto, "output", "", "")

	flag.StringVar(&o.Method, "d", "", "")
	flag.StringVar(&o.Method, "data", "", "")

	flag.Var(&o.Headers, "header", "")
	flag.Var(&o.Headers, "H", "")

	flag.StringVar(&o.Proxy, "x", "", "")
	flag.StringVar(&o.Proxy, "proxy", "", "")

	flag.IntVar(&o.Concurrency, "c", 25, "")
	flag.IntVar(&o.Concurrency, "concurrent", 25, "")

	flag.BoolVar(&o.Silent, "s", false, "")
	flag.BoolVar(&o.Silent, "silent", false, "")

	flag.BoolVar(&o.Verbose, "v", false, "")
	flag.BoolVar(&o.Verbose, "verbose", false, "")

	flag.BoolVar(&o.Version, "V", false, "")
	flag.BoolVar(&o.Version, "version", false, "")

	// Override help text
	flag.Usage = func() {
		showBanner()
		h := []string{
			"",
			"Usage:" + usage,
			"",
			"Options:" + options,
			"",
		}

		fmt.Fprint(os.Stderr, strings.Join(h, "\n"))
	}

	flag.Parse()
}

// ParseOptions will parse given args
func ParseOptions() *Options {
	// Show current version & exit
	if o.Version {
		showVersion()
	}

	// Show banner to user
	if !o.Silent {
		showBanner()
	}

	// Validate input options
	o.validate()

	return o
}
