package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func handleGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"message\": \"This is GET method!\"}"))
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := io.ReadAll(r.Body)
	w.Write([]byte(body))
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := io.ReadAll(r.Body)
	w.Write([]byte(body))
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := io.ReadAll(r.Body)
	w.Write([]byte(body))
}

func handlePatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := io.ReadAll(r.Body)
	w.Write([]byte(body))
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/get", handleGet)
	r.Post("/post", handlePost)
	r.Delete("/delete", handleDelete)
	r.Put("/put", handlePut)
	r.Patch("/patch", handlePatch)

	host := os.Args[1]

	origin := http.Server{
		Addr:    host,
		Handler: r,
	}
	log.Printf("Origin started at %s\n", host)
	if err := origin.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
