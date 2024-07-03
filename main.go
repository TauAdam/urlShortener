package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
)

func NewRedirectHandler(config map[string]string, fallback http.Handler) *RedirectHandler {
	return &RedirectHandler{
		config:   config,
		fallback: fallback,
	}
}

func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if dest, ok := h.config[path]; ok {
		http.Redirect(w, r, dest, http.StatusSeeOther)
		return
	}

	h.fallback.ServeHTTP(w, r)
}

func LoadConfigFromYaml(bytes []byte, cache *Cache) (map[string]string, error) {
	if cache.yamlConfig != nil {
		return cache.yamlConfig, nil
	}

	var slice []PathUrls
	err := yaml.Unmarshal(bytes, &slice)
	if err != nil {
		return nil, err
	}
	pathToUrls := make(map[string]string)
	for _, path := range slice {
		pathToUrls[path.Path] = path.Url
	}

	cache.yamlConfig = pathToUrls
	return pathToUrls, nil
}

func LoadConfigFromJson(bytes []byte, cache *Cache) (map[string]string, error) {
	if cache.jsonConfig != nil {
		return cache.jsonConfig, nil
	}

	var config JsonConfig
	err := json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	cache.jsonConfig = config.Config
	return config.Config, nil
}
func main() {
	cache := &Cache{}

	yml := `
- path: /rick
  url: https://www.youtube.com/watch?v=dQw4w9WgXcQ
- path: /google
  url: https://www.youtube.com/watch?v=dQw4w9WgXcQ`

	jsonConfig := `
{
	"config": {
		"/rick": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"/google": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	}
}`

	yamlConfig, err := LoadConfigFromYaml([]byte(yml), cache)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonConfigMap, err := LoadConfigFromJson([]byte(jsonConfig), cache)
	if err != nil {
		fmt.Println(err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", 301)
	})

	yamlHandler := NewRedirectHandler(yamlConfig, mux)
	jsonHandler := NewRedirectHandler(jsonConfigMap, mux)

	http.Handle("/yaml/", http.StripPrefix("/yaml/", yamlHandler))
	http.Handle("/json/", http.StripPrefix("/json/", jsonHandler))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening on port 8080")
}
