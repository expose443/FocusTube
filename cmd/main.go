package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.FormValue("code")))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
