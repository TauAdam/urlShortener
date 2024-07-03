package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
)

type PathUrls struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}

func main() {
	dictionary := map[string]string{
		"/rick":   "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"/google": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", 301)
	})
	handlerFunc := HandlerFunc(dictionary, mux)

	yml := `
- path: /rick
  url: https://www.youtube.com/watch?v=dQw4w9WgXcQ
- path: /google
  url: https://www.youtube.com/watch?v=dQw4w9WgXcQ`
	ymlHandlerFunc := YmlHandler([]byte(yml), mux)

	http.Handle("/", http.HandlerFunc(ymlHandlerFunc))
	err2 := http.ListenAndServe(":8010", nil)
	if err2 != nil {
		panic(err2)
	}

	http.Handle("/", http.HandlerFunc(handlerFunc))
	err := http.ListenAndServe(":8080", nil)
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

func YmlHandler(bytes []byte, fallback http.Handler) http.HandlerFunc {
	var slice []PathUrls
	err := yaml.Unmarshal(bytes, &slice)
	if err != nil {
		return nil
	}
	pathToUrls := make(map[string]string)
	for _, path := range slice {
		pathToUrls[path.Path] = path.Url
	}
	return HandlerFunc(pathToUrls, fallback)
}
