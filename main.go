package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	cdnURL := flag.String("url", "", "CDN URL to cache")
	output := flag.String("output", "", "Output folder")
	flag.Parse()

	if *cdnURL == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "Usage: esm-cache -url <CDN_URL> -output <OUTPUT_FOLDER>")
		os.Exit(1)
	}

	if err := run(*cdnURL, *output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cdnURL, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	cache := &moduleCache{
		outputDir: outputDir,
		modules:   make(map[string]string),
		contents:  make(map[string]string),
	}

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{cdnURL},
		Bundle:      true,
		Write:       false,
		Format:      api.FormatESModule,
		Platform:    api.PlatformBrowser,
		Plugins: []api.Plugin{
			{
				Name: "http-loader",
				Setup: func(build api.PluginBuild) {
					build.OnResolve(api.OnResolveOptions{Filter: ".*"},
						func(args api.OnResolveArgs) (api.OnResolveResult, error) {
							if strings.HasPrefix(args.Path, "http://") || strings.HasPrefix(args.Path, "https://") {
								return api.OnResolveResult{
									Path:      args.Path,
									Namespace: "http",
								}, nil
							}

							if args.Importer != "" {
								base, err := url.Parse(args.Importer)
								if err == nil && (base.Scheme == "http" || base.Scheme == "https") {
									resolved := resolveURL(base, args.Path)
									return api.OnResolveResult{
										Path:      resolved,
										Namespace: "http",
									}, nil
								}
							}

							return api.OnResolveResult{}, nil
						})

					build.OnLoad(api.OnLoadOptions{Filter: ".*", Namespace: "http"},
						func(args api.OnLoadArgs) (api.OnLoadResult, error) {
							if cached, ok := cache.contents[args.Path]; ok {
								return api.OnLoadResult{
									Contents: &cached,
									Loader:   api.LoaderJS,
								}, nil
							}

							content, err := downloadURL(args.Path)
							if err != nil {
								return api.OnLoadResult{}, err
							}

							cache.addModule(args.Path, content)

							return api.OnLoadResult{
								Contents: &content,
								Loader:   api.LoaderJS,
							}, nil
						})
				},
			},
		},
	})

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s\n", err.Text)
		}
		return fmt.Errorf("build failed with %d errors", len(result.Errors))
	}

	for moduleURL, content := range cache.contents {
		localPath := cache.modules[moduleURL]
		fullPath := filepath.Join(outputDir, localPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return cache.writeImportMap()
}

type moduleCache struct {
	outputDir string
	modules   map[string]string
	contents  map[string]string
}

func (c *moduleCache) addModule(moduleURL, content string) string {
	if localPath, exists := c.modules[moduleURL]; exists {
		return localPath
	}

	localPath := urlToPath(moduleURL)
	c.modules[moduleURL] = localPath
	c.contents[moduleURL] = content
	return localPath
}

func (c *moduleCache) writeImportMap() error {
	importMap := map[string]interface{}{
		"imports": c.modules,
	}

	data, err := json.MarshalIndent(importMap, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(c.outputDir, "importmap.json"), data, 0644)
}

func urlToPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		hash := sha256.Sum256([]byte(rawURL))
		return "invalid-" + hex.EncodeToString(hash[:8]) + ".js"
	}

	path := filepath.Join(u.Host, u.Path)

	if !strings.HasSuffix(path, ".js") && !strings.HasSuffix(path, ".mjs") {
		path = path + ".js"
	}

	return path
}

func downloadURL(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, rawURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func resolveURL(base *url.URL, ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(refURL).String()
}
