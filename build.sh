#!/bin/sh

set -e

rm -rf esm-vendor
go build -o esm-cache
./esm-cache

cat > public/index.html << 'HTMLSTART'
<!doctype html>
<html>
    <head>
        <meta charset="utf-8" />
        <title>ESM Vendor Test</title>
        <script type="importmap">
HTMLSTART

cat public/js/vendor/importmap.json >> public/index.html

cat >> public/index.html << 'HTMLEND'
        </script>
    </head>
    <body>
        <svg width="500" height="500"></svg>

        <script type="module">
            import * as d3 from "d3";
            import confetti from "confetti";

            const data = [30, 86, 168, 234, 155, 98];
            const svg = d3.select("svg");

            svg.selectAll("rect")
                .data(data)
                .join("rect")
                .attr("x", (d, i) => i * 80)
                .attr("y", (d) => 500 - d)
                .attr("width", 70)
                .attr("height", (d) => d)
                .attr("fill", "steelblue");

            confetti({
                particleCount: 100,
                spread: 70,
                origin: { y: 0.6 }
            });
        </script>
    </body>
</html>
HTMLEND

echo "Done! Serve with: python -m http.server 8000 -d public"
