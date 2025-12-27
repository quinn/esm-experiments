package main

import (
	"flag"
	"fmt"
	"os"

	"esm-cache/esm"
)

func main() {
	cdnURL := flag.String("url", "", "CDN URL to cache")
	output := flag.String("output", "", "Output folder")
	importName := flag.String("name", "", "Import name for top-level module")
	flag.Parse()

	if *cdnURL == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "Usage: esm-cache -url <CDN_URL> -output <OUTPUT_FOLDER>")
		os.Exit(1)
	}

	config := esm.Config{
		URL:        *cdnURL,
		OutputDir:  *output,
		ImportName: *importName,
	}

	if err := esm.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
