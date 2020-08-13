package crlfuzz

var (
	cookie   = "param=crlfuzz"
	injector = "Set-Cookie:" + cookie + ";"

	appendList = [...]string{
		"",
		"crlf",
		"?crlf=",
		"#",
	}

	escapeList = [...]string{
		"%0d",
		"%0a",
		"%0d%0a",
		"%23%0d",
		"%23%0a",
		"%23%0d%0a",
	}
)
