const esbuild = require("esbuild")

esbuild.build({
  entryPoints: ["src/index.ts"],
  format: "cjs",
  platform: "node",
  bundle: true,
  minify: true,
  outfile: "dist/index.js",
})
