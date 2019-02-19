package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var config *Config

func init() {
	file, err := os.Open("config/conf.json")
	if err != nil {
		fmt.Println("error:", err)
		panic(err)
	}
	decoder := json.NewDecoder(file)
	v := Config{}
	err = decoder.Decode(&v)
	if err != nil {
		fmt.Println("error:", err)
		panic(err)
	}
	config = &v
}

func GetConfig() *Config {
	return config
}

type Config struct {
	GCProjectID string `json:"gc_project_id"`
	ENV         string `json:"env"`
	Port        string `json:"port"`
}
