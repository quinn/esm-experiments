<!doctype html>
<html>
    <head>
        <meta charset="utf-8" />
        <title>ESM Vendor Test</title>
        <script type="importmap">{{ .importmap }}</script>
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
