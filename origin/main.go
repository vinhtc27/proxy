package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	host := os.Args[1]
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Printf("[origin server] - %s: %s\n", host, time.Now().Format("15:04:05"))
		_, _ = fmt.Fprintf(rw, "[%s] origin server response\n", host)
	})
	orgin := http.Server{
		Addr:    host,
		Handler: handler,
	}
	log.Printf("Origin started at %s\n", host)
	if err := orgin.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
