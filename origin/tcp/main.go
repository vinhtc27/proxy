package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	host := os.Args[1]
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, err := fmt.Fprintf(w, "[%s] origin server response\n", host)
		if err != nil {
			fmt.Printf("[%s] response write error : %v\n", host, err)
		}
	})
	origin := http.Server{
		Addr:    host,
		Handler: handler,
	}
	log.Printf("Origin started at %s\n", host)
	if err := origin.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
