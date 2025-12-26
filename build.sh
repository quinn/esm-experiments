#!/bin/sh

set -e

rm -rf cache
go build -o esm-cache
./esm-cache -url https://esm.sh/d3@7 -output cache -name d3

cat > test.html << 'HTMLSTART'
<!doctype html>
<html>
    <head>
        <meta charset="utf-8" />
        <title>D3 Test</title>
        <script type="importmap">
HTMLSTART

cat cache/importmap.json >> test.html

cat >> test.html << 'HTMLEND'
        </script>
    </head>
    <body>
        <svg width="500" height="500"></svg>

        <script type="module">
            import * as d3 from "d3";

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
        </script>
    </body>
</html>
HTMLEND

echo "Done! Serve with: python3 -m http.server 8000"
