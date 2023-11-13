package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxy/proxy/websocket"
	"proxy/ratelimit/token_bucket"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	Url     *url.URL
	Alive   bool
	Reverse *httputil.ReverseProxy
}

func NewServer(s string) *Server {
	url, err := url.Parse(s)
	if err != nil {
		log.Fatal(err)
	}

	transport := &http.Transport{
		MaxIdleConns:          100,              // Adjust based on expected load.
		MaxIdleConnsPerHost:   10,               // Limit idle connections per host.
		MaxConnsPerHost:       0,                // No limit on the total connections per host.
		ResponseHeaderTimeout: 2 * time.Second,  // Adjust based on desired response time.
		ExpectContinueTimeout: 1 * time.Second,  // Adjust based on desired behavior.
		IdleConnTimeout:       30 * time.Second, // Adjust based on desired connection reuse.
		Dial: (&net.Dialer{
			Timeout:   1 * time.Second,  // Adjust connection timeout as needed.
			KeepAlive: 30 * time.Second, // Adjust keep-alive time as needed.
		}).Dial,
	}

	reverse := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = url
		},
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Origin server error", http.StatusInternalServerError)
		},
	}

	return &Server{
		Url:     url,
		Alive:   true,
		Reverse: reverse,
	}
}

type ServerPool struct {
	servers []*Server
	index   int64
}

func (s *ServerPool) AddServer(server *Server) {
	s.servers = append(s.servers, server)
}

func (s *ServerPool) GetServer() *Server {
	atomic.AddInt64(&s.index, 1)
	//TODO: check overflow s.index
	return s.servers[s.index%int64(len(s.servers))]
}

func main() {
	var serversArg string
	var websocketArg string
	flag.StringVar(&serversArg, "servers", "", "Load balanced servers, use commas to separate")
	flag.StringVar(&websocketArg, "websocket", "", "Whether to use websocket")
	flag.Parse()
	if len(serversArg) == 0 {
		log.Fatal("Missing servers parameter")
	}

	servers := strings.Split(serversArg, ",")
	serverPool := ServerPool{index: -1}
	for _, s := range servers {
		serverPool.AddServer(NewServer(s))
	}

	host := "127.0.0.1:9090"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server := serverPool.GetServer()
		if server != nil {
			server.Reverse.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Ogirin server unavailable", http.StatusServiceUnavailable)
	})

	// For websocket proxy handler
	if len(websocketArg) > 0 {
		wp, err := websocket.NewProxy("ws://localhost:8312/", func(r *http.Request) error {
			// Permission to verify
			r.Header.Set("Cookie", "----")
			// Source of disguise
			r.Header.Set("Origin", "http://localhost:8312")
			return nil
		})
		if err != nil {
			log.Fatal()
		}
		
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wp.Proxy(w, r)
		})
	
	}

	proxy := http.Server{
		Addr:              host,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       30 * time.Second,
		// Handler:           handler,
		Handler: token_bucket.RequestThrottler(handler, 100),
	}

	log.Printf("Proxy started at %s\n", host)
	if err := proxy.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
