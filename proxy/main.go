package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxy/algorithm/sliding_window"
	"strings"
	"time"
)

type Server struct {
	URL          *url.URL
	Alive        bool
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	servers []*Server
	current int64
}

func (s *ServerPool) AddServer(server *Server) {
	s.servers = append(s.servers, server)
}

func (s *ServerPool) NextServerIndex() int64 {
	s.current++
	return s.current % int64(len(s.servers))
}

func (s *ServerPool) GetNextServer() *Server {
	next := s.NextServerIndex()
	return s.servers[next]
}

func main() {
	var serversArg string
	flag.StringVar(&serversArg, "servers", "", "Load balanced servers, use commas to separate")
	flag.Parse()

	if len(serversArg) == 0 {
		log.Fatal("Missing servers parameter")
	}

	servers := strings.Split(serversArg, ",")

	serverPool := ServerPool{current: -1}
	for _, s := range servers {
		serverUrl, err := url.Parse(s)

		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		serverPool.AddServer(&Server{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}

	host := "127.0.0.1:9090"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer := serverPool.GetNextServer()

		if peer != nil {
			peer.ReverseProxy.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	})
	proxy := http.Server{
		Addr:              host,
		ReadTimeout:       time.Second,
		WriteTimeout:      time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       10 * time.Second,
		Handler:           sliding_window.RequestThrottler(handler, 10000),
	}

	log.Printf("Proxy started at %s\n", host)
	if err := proxy.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
