#!/bin/sh

set -e

rm -rf public/js/vendor
rm public/index.html
go build -o esm-vendor
./esm-vendor

echo "Done! Serve with: python -m http.server 8000 -d public"
