package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	proxyScheme = "https"
	proxyHost   = "oauth2.googleapis.com"
)

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	http.HandleFunc("/", reqIDMiddleware1(proxyRequest(config)))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server listening on :" + port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("server error: %v\n", err)
	}
}

func proxyRequest(c *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		proxyUrl := fmt.Sprintf("%s://%s%s", proxyScheme, proxyHost, req.RequestURI)
		err := req.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		values := cloneURLValues(req.Form)
		values.Add("client_secret", c.ClientSecret)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		proxyReq, err := http.NewRequestWithContext(ctx, req.Method, proxyUrl, strings.NewReader(values.Encode()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			GetRequestLogger(req).Errorln(err)
			return
		}

		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			if prior, ok := proxyReq.Header["X-Forwarded-For"]; ok {
				clientIP = strings.Join(prior, ", ") + ", " + clientIP
			}
			proxyReq.Header.Set("X-Forwarded-For", clientIP)
		}

		// We may want to filter some headers, otherwise we could just use a shallow copy
		// proxyReq.Header = req.Header
		proxyReq.Header = make(http.Header)
		for h, val := range req.Header {
			proxyReq.Header[h] = val
		}

		// step 2
		httpClient := http.Client{
			Transport: http.DefaultTransport,
			Timeout:   20 * time.Second,
		}
		res, err := httpClient.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			GetRequestLogger(req).Errorln(err)
			return
		}
		if res != nil && res.Body != nil {
			defer func() {
				io.Copy(ioutil.Discard, res.Body)
				res.Body.Close()
			}()
		}

		// step 3
		for key, value := range res.Header {
			for _, v := range value {
				w.Header().Add(key, v)
			}
		}

		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
	}
}

func cloneURLValues(v url.Values) url.Values {
	v2 := make(url.Values, len(v))
	for k, vv := range v {
		v2[k] = append([]string(nil), vv...)
	}
	return v2
}
