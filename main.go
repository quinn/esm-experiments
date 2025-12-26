package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	url := flag.String("url", "", "CDN URL to cache")
	output := flag.String("output", "", "Output folder")
	flag.Parse()

	if *url == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "Usage: esm-cache -url <CDN_URL> -output <OUTPUT_FOLDER>")
		os.Exit(1)
	}

	if err := run(*url, *output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cdnURL, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	resp, err := http.Get(cdnURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	entryFile := filepath.Join(outputDir, "entry.js")
	if err := os.WriteFile(entryFile, body, 0644); err != nil {
		return err
	}

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{entryFile},
		Bundle:      true,
		Outdir:      outputDir,
		Format:      api.FormatESModule,
		Write:       true,
		Splitting:   true,
		Metafile:    true,
		Platform:    api.PlatformBrowser,
	})

	if len(result.Errors) > 0 {
		return fmt.Errorf("build errors: %v", result.Errors)
	}

	importMap := map[string]interface{}{
		"imports": map[string]string{},
	}

	if result.Metafile != "" {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(result.Metafile), &meta); err == nil {
			outputs := meta["outputs"].(map[string]interface{})
			for outputPath := range outputs {
				relPath := strings.TrimPrefix(outputPath, outputDir+"/")
				if relPath != "entry.js" {
					importMap["imports"].(map[string]string)[cdnURL] = relPath
				}
			}
		}
	}

	for _, file := range result.OutputFiles {
		name := filepath.Base(file.Path)
		if name != "entry.js" {
			importMap["imports"].(map[string]string)[cdnURL] = name
		}
	}

	mapPath := filepath.Join(outputDir, "importmap.json")
	mapData, err := json.MarshalIndent(importMap, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(mapPath, mapData, 0644)
}
