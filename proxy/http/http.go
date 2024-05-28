package http

import (
	"crypto/tls"
	"encoding/base64"
	"go-woof/utils"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Http struct {
	conf utils.HttpConf
}

func NewHttp(conf utils.HttpConf) *Http {
	h := &Http{
		conf: conf,
	}
	return h
}

func (h *Http) AuthCheck(auth string) bool {
	userPass := url.UserPassword(h.conf.Username, h.conf.Password)
	auth = strings.ReplaceAll(auth, "Basic ", "")
	if auth != base64.StdEncoding.EncodeToString([]byte(userPass.String())) {
		return false
	}
	return true
}

func (h *Http) httpHandle(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Proxy-Authorization")
	if authorization != "" && !h.AuthCheck(authorization) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	if r.Method == http.MethodConnect {
		h.handleHttps(w, r)
	} else {
		h.handleHttp(w, r)
	}
}

func (h *Http) handleHttp(w http.ResponseWriter, r *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *Http) handleHttps(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 60*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func (h *Http) Run() {
	cert, err := GenCertificate()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:      h.conf.ServerAddr,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			h.httpHandle(writer, request)
		}),
	}
	log.Println("start listen addr", h.conf.ServerAddr)
	log.Fatal(server.ListenAndServe())
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		if k == "Proxy-Authorization" {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
