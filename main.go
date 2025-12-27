package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"esm-cache/esm"
)

func main() {
	configFile := flag.String("config", "esm-cache.config.json", "Config file path (config file mode)")
	flag.Parse()

	var config esm.Config

	cfg, err := loadConfigFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
		os.Exit(1)
	}
	config = cfg

	if err := esm.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func loadConfigFile(path string) (esm.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return esm.Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config esm.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return esm.Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
