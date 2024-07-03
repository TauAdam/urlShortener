package main

import (
	"fmt"
	"net/http"
)

func main() {
	dictionary := map[string]string{
		"/rick":   "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"/google": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", 301)
	})
	handler := HandlerFunc(dictionary, mux)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		return
	}
	fmt.Println("Listening on port 8080")

}
func HandlerFunc(dict map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if dest, ok := dict[path]; ok {
			http.Redirect(w, r, dest, http.StatusSeeOther)
			return
		}

		fallback.ServeHTTP(w, r)
	}
}
