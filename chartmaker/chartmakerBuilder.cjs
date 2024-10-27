const esbuild = require("esbuild")

esbuild.build({
  entryPoints: ["src/server.ts"],
  outdir: "dist",
  format: "cjs",
  platform: "node",
  external: ["@napi-rs/canvas"],
  bundle: true,
  minify: true,
})
