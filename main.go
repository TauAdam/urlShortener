package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
)

type RedirectHandler struct {
	config   map[string]string
	fallback http.Handler
}

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

func LoadConfigFromYaml(bytes []byte) (map[string]string, error) {
	var slice []PathUrls
	err := yaml.Unmarshal(bytes, &slice)
	if err != nil {
		return nil, err
	}
	pathToUrls := make(map[string]string)
	for _, path := range slice {
		pathToUrls[path.Path] = path.Url
	}
	return pathToUrls, nil
}

type PathUrls struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}

type JsonConfig struct {
	Config map[string]string `json:"config"`
}

func LoadConfigFromJson(bytes []byte) (map[string]string, error) {
	var config JsonConfig
	err := json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return config.Config, nil
}

func main() {
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

	yamlConfig, err := LoadConfigFromYaml([]byte(yml))
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonConfigMap, err := LoadConfigFromJson([]byte(jsonConfig))
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
