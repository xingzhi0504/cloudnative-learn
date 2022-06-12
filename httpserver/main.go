package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
)

const (
	version = "1.0.0"
)

func home(w http.ResponseWriter, r *http.Request) {
	ver, ok := os.LookupEnv("VERSION")
	if !ok {
		if err := os.Setenv("VERSION", version); err != nil {
			log.Println(fmt.Errorf("set env var VERSION faild, err: %v", err))
		} else {
			ver = os.Getenv("VERSION")
		}
	}

	// 1、接收客户端 request，并将 request 中带的 header 写入 response header
	for k, v := range r.Header {
		for _, vv := range v {
			w.Header().Set(k, vv)
		}
	}
	// 2、读取当前系统的环境变量中的 VERSION 配置，并写入 response header
	w.Header().Set("VERSION", ver)

	// 3、Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
	w.WriteHeader(http.StatusOK)
	log.Println(fmt.Sprintf("请求uri: %s, ip: %s, 返回码：%d", r.URL.Path, getClientIp(r), http.StatusOK))
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("working"))
}

func getClientIp(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	if ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0]); len(ip) > 0 {
		return ip
	}

	if ip := strings.TrimSpace(r.Header.Get("X-Real-Ip")); len(ip) > 0 {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

/*
1、接收客户端 request，并将 request 中带的 header 写入 response header
2、读取当前系统的环境变量中的 VERSION 配置，并写入 response header
3、Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
4、当访问 localhost/healthz 时，应返回 200
*/
func main() {

	mux := http.NewServeMux()
	// main route
	mux.HandleFunc("/", home)
	mux.HandleFunc("/healthz", healthz)

	// debug for pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("started http server faild, err: %v", err)
	}
}
