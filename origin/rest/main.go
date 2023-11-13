package main

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func handleGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"message\": \"This is GET method!\"}"))
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.Write([]byte(body))
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.Write([]byte(body))
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.Write([]byte(body))
}

func handleHead(w http.ResponseWriter, r *http.Request) {
	// res = r.Response.Request.Header.Get("")
	// body, _ := io.ReadAll(r.Body)
	// b := new(bytes.Buffer)
	// for key, value := range r.Header {
	// 	fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	// }
	// w.Write([]byte(b.String()))
}

func handlePatch(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.Write([]byte(body))
}

func handleOption(w http.ResponseWriter, r *http.Request) {
	// body, _ := io.ReadAll(r.Body)
	// w.Write([]byte(body))
}

func handleTrace(w http.ResponseWriter, r *http.Request) {
	// body, _ := io.ReadAll(r.Body)
	// w.Write([]byte(body))
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
	r.Head("/head", handleHead)
	r.Options("/option", handleOption)
	r.Trace("/trace", handleTrace)

	http.ListenAndServe(":8212", r)
}
