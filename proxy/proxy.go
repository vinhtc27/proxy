package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"proxy/ratelimit/fixed_window"
	"proxy/ratelimit/sliding_log"
	"proxy/ratelimit/sliding_window"
	"proxy/ratelimit/token_bucket"
	"sync/atomic"
	"time"

	"golang.org/x/net/http2"
)

type Config struct {
	Address              string   `json:"address"`
	EnableRateLimit      bool     `json:"enableRateLimit"`
	RateLimitType        string   `json:"rateLimitType"`
	RatePerSecond        int64    `json:"ratePerSecond"`
	LoadBalanceEndpoints []string `json:"loadBalanceEndpoints"`
}

type ProxyConfig struct {
	Config     *Config
	ServerPool *ServerPool
}

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

	err = http2.ConfigureTransport(transport)
	if err != nil {
		log.Fatal(err)
	}

	reverse := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = url.Scheme
			req.URL.Host = url.Host
		},
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, fmt.Sprintf("Origin server error %s", err), http.StatusInternalServerError)
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
	var proxyConfig = new(ProxyConfig)
	jsonFile, err := os.Open("./config.json")
	if err != nil {
		log.Panic("Error: os.Open(\"./config.json\")", err.Error())
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			log.Panic("Error: jsonFile.Close()", err.Error())
		}
	}(jsonFile)

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Panic("Error: io.ReadAll(jsonFile)", err.Error())
	}
	err = json.Unmarshal(byteValue, &proxyConfig.Config)
	if err != nil {
		log.Panic("Error: json.Unmarshal", err.Error())
	}

	if proxyConfig.Config.Address == "" {
		log.Panic("Error: config.json must have address")
	}
	if len(proxyConfig.Config.LoadBalanceEndpoints) == 0 {
		log.Panic("Error: config.json must have loadBalanceEndpoints")
	}
	if proxyConfig.Config.EnableRateLimit {
		if proxyConfig.Config.RateLimitType == "" {
			log.Panic("Error: config.json must have rateLimitType")
		}
		if proxyConfig.Config.RatePerSecond == 0 {
			log.Panic("Error: config.json must have ratePerSecond")
		}
	}

	var handler http.HandlerFunc
	if len(proxyConfig.Config.LoadBalanceEndpoints) >= 2 {
		proxyConfig.ServerPool = &ServerPool{index: -1}
		for _, s := range proxyConfig.Config.LoadBalanceEndpoints {
			proxyConfig.ServerPool.AddServer(NewServer(s))
		}
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			server := proxyConfig.ServerPool.GetServer()
			if server != nil {
				server.Reverse.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Origin server unavailable", http.StatusServiceUnavailable)
		})
	} else {
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			server := NewServer(proxyConfig.Config.LoadBalanceEndpoints[0])
			if server != nil {
				server.Reverse.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Origin server unavailable", http.StatusServiceUnavailable)
		})
	}

	var httpServer http.Server
	if proxyConfig.Config.EnableRateLimit {
		var rateLimitHandler http.Handler
		switch proxyConfig.Config.RateLimitType {
		case "fixed_window":
			rateLimitHandler = fixed_window.RequestThrottler(handler, proxyConfig.Config.RatePerSecond)
		case "sliding_log":
			rateLimitHandler = sliding_log.RequestThrottler(handler, proxyConfig.Config.RatePerSecond)
		case "sliding_window":
			rateLimitHandler = sliding_window.RequestThrottler(handler, proxyConfig.Config.RatePerSecond)
		default:
			rateLimitHandler = token_bucket.RequestThrottler(handler, proxyConfig.Config.RatePerSecond)
		}
		httpServer = http.Server{
			Addr:              proxyConfig.Config.Address,
			ReadTimeout:       1 * time.Second,
			WriteTimeout:      1 * time.Second,
			ReadHeaderTimeout: 2 * time.Second,
			IdleTimeout:       30 * time.Second,
			Handler:           rateLimitHandler,
		}
	} else {
		httpServer = http.Server{
			Addr:              proxyConfig.Config.Address,
			ReadTimeout:       1 * time.Second,
			WriteTimeout:      1 * time.Second,
			ReadHeaderTimeout: 2 * time.Second,
			IdleTimeout:       30 * time.Second,
			Handler:           handler,
		}
	}

	fmt.Println(proxyConfig.Config.LoadBalanceEndpoints)
	log.Printf("Proxy started at %s\n", proxyConfig.Config.Address)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
