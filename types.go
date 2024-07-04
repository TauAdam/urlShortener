package main

import "net/http"

type PathUrls struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}

type JsonConfig struct {
	Config map[string]string `json:"config"`
}
type RedirectHandler struct {
	config   map[string]string
	fallback http.Handler
}

type RedirectRequest struct {
	Path string
	Url  string
}
type Cache struct {
	yamlConfig map[string]string
	jsonConfig map[string]string
	tomlConfig map[string]string
}
