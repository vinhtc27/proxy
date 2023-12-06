package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	host := os.Getenv("HOST")
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, err := fmt.Fprintf(w, "[%s] TCP response\n", host)
		if err != nil {
			fmt.Printf("[%s] response write error : %v\n", host, err)
		}
		fmt.Printf("TCP response at %v from %s\n", time.Now().Format("2006-01-02 15:04:05"), host)
	})
	origin := http.Server{
		Addr:    host,
		Handler: handler,
	}
	log.Printf("TCP started at %s\n", host)
	if err := origin.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
