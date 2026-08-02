package request

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// dialTimeout is used for establishing connections and for the per-request
// read/write deadlines.
const dialTimeout = time.Second * 30

// RawRequest performs an HTTP request whose request target may contain raw
// CRLF control characters. Go's net/http refuses to write control characters
// in the request URI (see net/http/request.go, stringContainsCTLByte), so the
// request is assembled byte-for-byte and sent over a raw connection instead.
func RawRequest(method, rawurl, data string, headers []string, proxy string) (*http.Response, error) {
	scheme, host, target, err := splitTarget(rawurl)
	if err != nil {
		return nil, err
	}

	absolute := viaPlainProxy(proxy, scheme)
	proxyAuth := proxyAuthorization(proxy)

	payload := buildRequest(method, scheme, host, target, data, headers, absolute, proxyAuth)

	conn, err := dial(scheme, host, proxy)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(dialTimeout)); err != nil {
		return nil, err
	}

	if _, err := conn.Write(payload); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: method})
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	return resp, nil
}

// splitTarget extracts the scheme, host and raw request target from a URL that
// may contain control characters, which url.Parse would otherwise reject.
func splitTarget(rawurl string) (scheme, host, target string, err error) {
	i := strings.Index(rawurl, "://")
	if i == -1 {
		return "", "", "", fmt.Errorf("invalid target URL: %q", rawurl)
	}
	scheme = rawurl[:i]
	rest := rawurl[i+3:]

	j := strings.IndexAny(rest, "/?#")
	if j == -1 {
		return scheme, rest, "/", nil
	}

	return scheme, rest[:j], rest[j:], nil
}

// viaPlainProxy reports whether the request line must be written in
// absolute-form, i.e. when tunneling through an HTTP-family proxy for a
// non-TLS target. RFC 7230 §5.3.2 requires absolute-form for forward proxies;
// SOCKS5 proxies carry the target address in the handshake instead.
func viaPlainProxy(proxy, scheme string) bool {
	if proxy == "" || scheme == "https" {
		return false
	}

	u, err := url.Parse(proxy)
	if err != nil {
		return false
	}

	switch u.Scheme {
	case "http", "https":
		return true
	default:
		return false
	}
}

// buildRequest assembles the raw HTTP request bytes, keeping any injected CRLF
// sequences in the request target intact.
func buildRequest(method, scheme, host, target, data string, headers []string, absolute bool, proxyAuth string) []byte {
	var b bytes.Buffer

	ruri := target
	if absolute {
		ruri = scheme + "://" + host + target
	}

	fmt.Fprintf(&b, "%s %s HTTP/1.1\r\n", method, ruri)
	fmt.Fprintf(&b, "Host: %s\r\n", host)

	hasUserAgent := false
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(parts[0]), "User-Agent") {
			hasUserAgent = true
		}
	}

	if !hasUserAgent {
		fmt.Fprintf(&b, "User-Agent: Go-http-client/1.1\r\n")
	}

	if absolute && proxyAuth != "" {
		fmt.Fprintf(&b, "Proxy-Authorization: %s\r\n", proxyAuth)
	}

	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)

		if len(parts) != 2 {
			continue
		}

		fmt.Fprintf(&b, "%s: %s\r\n", parts[0], parts[1])
	}

	if data != "" {
		fmt.Fprintf(&b, "Content-Length: %d\r\n", len(data))
	}
	fmt.Fprintf(&b, "Connection: close\r\n\r\n")
	if data != "" {
		b.WriteString(data)
	}

	return b.Bytes()
}

// dial opens a connection to the target, handling TLS and proxies.
func dial(scheme, host, proxy string) (net.Conn, error) {
	d := &net.Dialer{Timeout: dialTimeout, KeepAlive: time.Second}

	if proxy != "" {
		return dialViaProxy(d, scheme, host, proxy)
	}

	conn, err := d.Dial("tcp", hostPort(host, defaultPort(scheme)))
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(dialTimeout))

	if scheme == "https" {
		tc := tls.Client(conn, &tls.Config{InsecureSkipVerify: true, ServerName: sniHost(host)})
		if err := tc.Handshake(); err != nil {
			_ = conn.Close()
			return nil, err
		}
		return tc, nil
	}

	return conn, nil
}

// dialViaProxy opens a connection through the given proxy.
func dialViaProxy(d *net.Dialer, scheme, host, proxy string) (net.Conn, error) {
	u, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		port := "80"
		if u.Scheme == "https" {
			port = "443"
		}

		conn, err := d.Dial("tcp", hostPort(u.Host, port))
		if err != nil {
			return nil, err
		}
		_ = conn.SetDeadline(time.Now().Add(dialTimeout))

		if u.Scheme == "https" {
			tc := tls.Client(conn, &tls.Config{InsecureSkipVerify: true, ServerName: sniHost(u.Host)})
			if err := tc.Handshake(); err != nil {
				_ = conn.Close()
				return nil, err
			}
			conn = tc
		}

		if scheme == "https" {
			return tunnelThroughProxy(conn, host, proxyAuthorization(proxy))
		}
		return conn, nil

	case "socks5", "socks5h":
		return dialSocks5(d, u, scheme, host)

	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %q", u.Scheme)
	}
}

// tunnelThroughProxy establishes a CONNECT tunnel for a TLS target.
func tunnelThroughProxy(conn net.Conn, host, proxyAuth string) (net.Conn, error) {
	addr := hostPort(host, "443")

	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
	if proxyAuth != "" {
		req += "Proxy-Authorization: " + proxyAuth + "\r\n"
	}
	req += "\r\n"

	if _, err := conn.Write([]byte(req)); err != nil {
		_ = conn.Close()
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodConnect})
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_ = conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: %s", resp.Status)
	}

	tc := tls.Client(conn, &tls.Config{InsecureSkipVerify: true, ServerName: sniHost(host)})
	if err := tc.Handshake(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return tc, nil
}

// dialSocks5 establishes a connection through a SOCKS5 proxy (RFC 1928).
func dialSocks5(d *net.Dialer, u *url.URL, scheme, host string) (net.Conn, error) {
	conn, err := d.Dial("tcp", hostPort(u.Host, "1080"))
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(dialTimeout))

	methods := []byte{0x00}
	if u.User != nil {
		methods = append(methods, 0x02)
	}

	greeting := []byte{0x05, byte(len(methods))}
	greeting = append(greeting, methods...)
	if _, err := conn.Write(greeting); err != nil {
		_ = conn.Close()
		return nil, err
	}

	var resp [2]byte
	if _, err := io.ReadFull(conn, resp[:]); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if resp[0] != 0x05 {
		_ = conn.Close()
		return nil, fmt.Errorf("invalid SOCKS5 response")
	}

	switch resp[1] {
	case 0x00: // no authentication required
	case 0x02: // username/password
		user := u.User.Username()
		pass, _ := u.User.Password()

		auth := []byte{0x01, byte(len(user))}
		auth = append(auth, user...)
		auth = append(auth, byte(len(pass)))
		auth = append(auth, pass...)
		if _, err := conn.Write(auth); err != nil {
			_ = conn.Close()
			return nil, err
		}

		var authResp [2]byte
		if _, err := io.ReadFull(conn, authResp[:]); err != nil {
			_ = conn.Close()
			return nil, err
		}
		if authResp[1] != 0x00 {
			_ = conn.Close()
			return nil, fmt.Errorf("SOCKS5 authentication failed")
		}
	default:
		_ = conn.Close()
		return nil, fmt.Errorf("SOCKS5 server does not support our auth methods")
	}

	addr := hostPort(host, defaultPort(scheme))
	hostname, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	var port uint16
	fmt.Sscanf(portStr, "%d", &port)

	req := []byte{0x05, 0x01, 0x00}
	if ip := net.ParseIP(hostname); ip != nil {
		if v4 := ip.To4(); v4 != nil {
			req = append(req, 0x01)
			req = append(req, v4...)
		} else {
			req = append(req, 0x04)
			req = append(req, ip.To16()...)
		}
	} else {
		req = append(req, 0x03, byte(len(hostname)))
		req = append(req, hostname...)
	}
	req = append(req, byte(port>>8), byte(port))

	if _, err := conn.Write(req); err != nil {
		_ = conn.Close()
		return nil, err
	}

	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if header[0] != 0x05 || header[1] != 0x00 {
		_ = conn.Close()
		return nil, fmt.Errorf("SOCKS5 connect failed (REP=%d)", header[1])
	}

	switch header[3] {
	case 0x01:
		if _, err := io.ReadFull(conn, make([]byte, 4)); err != nil {
			_ = conn.Close()
			return nil, err
		}
	case 0x03:
		var length [1]byte
		if _, err := io.ReadFull(conn, length[:]); err != nil {
			_ = conn.Close()
			return nil, err
		}
		if _, err := io.ReadFull(conn, make([]byte, length[0])); err != nil {
			_ = conn.Close()
			return nil, err
		}
	case 0x04:
		if _, err := io.ReadFull(conn, make([]byte, 16)); err != nil {
			_ = conn.Close()
			return nil, err
		}
	default:
		_ = conn.Close()
		return nil, fmt.Errorf("SOCKS5 unsupported address type %d", header[3])
	}

	if _, err := io.ReadFull(conn, make([]byte, 2)); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}

// hostPort returns the host with a default port appended when missing.
func hostPort(host, defaultPort string) string {
	if strings.HasPrefix(host, "[") {
		if strings.HasSuffix(host, "]") {
			return host + ":" + defaultPort
		}
		return host
	}
	if strings.Contains(host, ":") {
		return host
	}
	return host + ":" + defaultPort
}

// sniHost returns the host portion of a host[:port] string for use as the TLS
// ServerName (SNI), stripping any port and IPv6 brackets.
func sniHost(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.Trim(host, "[]")
}

// proxyAuthorization returns the Proxy-Authorization header value for an
// http(s) proxy URL that carries userinfo, or "" when no credentials are
// present or the proxy is not an HTTP-family proxy.
func proxyAuthorization(proxy string) string {
	if proxy == "" {
		return ""
	}

	u, err := url.Parse(proxy)
	if err != nil {
		return ""
	}

	switch u.Scheme {
	case "http", "https":
	default:
		return ""
	}

	if u.User == nil || u.User.Username() == "" {
		return ""
	}

	pass, _ := u.User.Password()
	token := u.User.Username() + ":" + pass
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(token))
}

// defaultPort returns the default port for the given scheme.
func defaultPort(scheme string) string {
	if scheme == "https" {
		return "443"
	}
	return "80"
}
