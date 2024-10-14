package main

import (
	"log"
	"net/http"
)

func main() {
	http.Handle("/assets/", http.FileServer(http.Dir("./www/")))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./www/index.html")
	}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
