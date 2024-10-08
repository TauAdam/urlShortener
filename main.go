package main

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
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

	var slice []struct {
		Path string
		Url  string
	}
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

	var config struct {
		Config map[string]string
	}
	err := json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	cache.jsonConfig = config.Config
	return config.Config, nil
}

func LoadConfigFromTOML(bytes []byte, cache *Cache) (map[string]string, error) {
	var config map[string]map[string]string
	err := toml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	pathToUrls := make(map[string]string)
	for _, m := range config {
		for path, url := range m {
			pathToUrls[path] = url
		}
	}

	cache.tomlConfig = pathToUrls
	return pathToUrls, nil
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

	http.HandleFunc("/yaml/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[5:]
		if dest, ok := yamlConfig[path]; ok {
			http.Redirect(w, r, dest, http.StatusSeeOther)
			return
		}
		http.Error(w, "Alias not found", http.StatusNotFound)
	})

	http.HandleFunc("/json/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[5:]
		if dest, ok := jsonConfigMap[path]; ok {
			http.Redirect(w, r, dest, http.StatusSeeOther)
			return
		}
		http.Error(w, "Alias not found", http.StatusNotFound)
	})

	http.HandleFunc("/api/config/yaml", func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(yamlConfig)
		if err != nil {
			return
		}
	})

	http.HandleFunc("/api/config/json", func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(jsonConfigMap)
		if err != nil {
			return
		}
	})

	http.HandleFunc("/api/map", func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(cache.jsonConfig)
		if err != nil {
			log.Fatalln(err)
		}
	})
	http.HandleFunc("/api/config/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}

		var req RedirectRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		yamlConfig[req.Path] = req.Url
		jsonConfigMap[req.Path] = req.Url

		w.WriteHeader(http.StatusCreated)
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Listening on port 8080")

	err = saveMapToFile(yamlConfig, "yaml_config.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = saveMapToFile(jsonConfigMap, "json_config.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Maps saved to files")
}

func saveMapToFile(mapData map[string]string, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	for key, value := range mapData {
		_, err := f.WriteString(fmt.Sprintf("%s - %s\n", key, value))
		if err != nil {
			return err
		}
	}

	ymlBytes, err := yaml.Marshal([]struct {
		Path string
		Url  string
	}{})
	for path, url := range mapData {
		ymlBytes = append(ymlBytes, []byte(fmt.Sprintf("- path: %s\n  url: %s\n", path, url))...)
	}
	err = os.WriteFile("yaml_config.yaml", ymlBytes, 0644)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(struct {
		Config map[string]string
	}{mapData})
	err = os.WriteFile("json_config.json", jsonBytes, 0644)
	if err != nil {
		return err
	}

	for path, url := range mapData {
		_, err := f.WriteString(fmt.Sprintf("[%s]\nurl = \"%s\"\n\n", path, url))
		if err != nil {
			return err
		}
	}

	return nil
}

