package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

var stripHeaders = map[string]bool{
	"host":                 true,
	"connection":           true,
	"keep-alive":           true,
	"proxy-authenticate":   true,
	"proxy-authorization":  true,
	"te":                   true,
	"trailer":              true,
	"transfer-encoding":    true,
	"upgrade":              true,
	"forwarded":            true,
	"x-forwarded-host":     true,
	"x-forwarded-proto":    true,
	"x-forwarded-port":     true,
}

func main() {
	target := os.Getenv("TARGET_DOMAIN")
	if target == "" {
		log.Fatal("TARGET_DOMAIN is not set")
	}

	target = strings.TrimRight(target, "/")

	remote, err := url.Parse(target)
	if err != nil {
		log.Fatal("Invalid TARGET_DOMAIN:", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.Host = remote.Host

		// Strip unwanted headers
		for k := range req.Header {
			lk := strings.ToLower(k)
			if stripHeaders[lk] || strings.HasPrefix(lk, "x-vercel-") {
				req.Header.Del(k)
			}
		}

		// Forward client IP if present
		if ip := req.Header.Get("X-Real-Ip"); ip != "" {
			req.Header.Set("X-Forwarded-For", ip)
		}
	}

	server := &http.Server{
		Addr: ":3000",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		}),
	}

	log.Println("Proxy running on :3000")
	log.Fatal(server.ListenAndServe())
}
