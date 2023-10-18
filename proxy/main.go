package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	tb "proxy/algorithm/token_bucket"
	"time"
)

func main() {

	// define origin server URL
	originServerURL, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		log.Fatal("invalid origin server URL")
	}

	proxy := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Printf("[proxy server] received request at: %s\n", time.Now())

		// set req Host, URL and Request URI to forward a request to the origin server
		req.Host = originServerURL.Host
		req.URL.Host = originServerURL.Host
		req.URL.Scheme = originServerURL.Scheme
		req.RequestURI = ""

		// send a request to the origin server
		originServerResponse, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(rw, err)
			return
		}

		rw.WriteHeader(http.StatusOK)
		io.Copy(rw, originServerResponse.Body)
	})

	log.Fatal(http.ListenAndServe("127.0.0.1:9090",
		tb.ReqTokenBucket(proxy, 5),
	))

}
