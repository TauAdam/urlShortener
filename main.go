package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", 301)
	})
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
	fmt.Println("Listening on port 8080")
}
