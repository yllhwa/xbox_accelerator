package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("mkcert.exe", "-install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	cmd = exec.Command("mkcert.exe", "-cert-file", "server.crt", "-key-file", "server.pem", "*.xboxlive.cn", "*.xboxlive.com")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	// file, _ := os.OpenFile("C:\\windows\\system32\\drivers\\etc\\hosts", os.O_RDONLY|os.O_APPEND, 0666)
	cmd = exec.Command("copy", "C:\\windows\\system32\\drivers\\etc\\hosts", "C:\\windows\\system32\\drivers\\etc\\hosts.goback")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", ServeHTTP)
		http.ListenAndServe(":80", mux)
	}()
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", ServeHTTP)
		http.ListenAndServeTLS(":443", "server.crt", "server.pem", mux)
	}()
	fmt.Println("服务器开始运行")
	select {}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	apath := a.EscapedPath()
	bpath := b.EscapedPath()
	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		req.Host = target.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	remote, err := url.Parse("http://" + "assets1.xboxlive.cn/")
	if err != nil {
		panic(err)
	}
	proxy := NewReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}
